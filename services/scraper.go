package services

import (
	"context"
	"fmt"
	"log"
	"pricemap-go/database"
	"pricemap-go/models"
	"pricemap-go/parsers"
	"sync"
	"time"
)

type ScraperService struct {
	parsers        []parsers.Parser
	factorsService *FactorsService
	metricsService *MetricsService
	cacheService   *CacheService
}

func NewScraperService() *ScraperService {
	return &ScraperService{
		parsers: []parsers.Parser{
			// Open Data Sources (most reliable, no blocking)
			parsers.NewNYCOpenDataParser(),    // NYC Open Data - Property Sales
			parsers.NewLondonOpenDataParser(), // London Data Store - House Prices
			parsers.NewBerlinOpenDataParser(), // Berlin Open Data - Real Estate
			parsers.NewParisOpenDataParser(),  // Paris Open Data - Rent Control
			parsers.NewTokyoOpenDataParser(),  // Tokyo Open Data - Property Prices
			parsers.NewSydneyOpenDataParser(), // Sydney Open Data - Property Sales
			parsers.NewMoscowOpenDataParser(), // Moscow Open Data - Real Estate

			// Commercial sites (may have blocking)
			parsers.NewCianParser(),      // Russia - 30+ cities, sale & rent
			parsers.NewRightmoveParser(), // UK - 25+ cities, sale & rent
			parsers.NewZillowParser(),    // USA - 30+ cities, sale & rent
			parsers.NewIdealistaParser(), // Spain - 20+ cities, sale & rent
		},
		factorsService: NewFactorsService(),
		metricsService: NewMetricsService(),
		cacheService:   NewCacheService(1 * time.Hour), // 1 hour TTL
	}
}

// ScrapeAll starts parsing all sources sequentially
func (ss *ScraperService) ScrapeAll(ctx context.Context) error {
	log.Println("Starting scraping process...")

	for _, parser := range ss.parsers {
		if err := ss.scrapeSource(ctx, parser); err != nil {
			log.Printf("Error scraping %s: %v", parser.Name(), err)
			continue
		}
	}

	log.Println("Scraping process completed")
	return nil
}

// ScrapeAllConcurrent starts parsing all sources concurrently with worker pool
func (ss *ScraperService) ScrapeAllConcurrent(ctx context.Context, workers int) error {
	log.Printf("Starting concurrent scraping with %d workers...", workers)

	// Channel for parser jobs
	jobs := make(chan parsers.Parser, len(ss.parsers))

	// Channel for results
	type result struct {
		parser string
		err    error
	}
	results := make(chan result, len(ss.parsers))

	// Start worker pool
	var wg sync.WaitGroup
	for w := 1; w <= workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for parser := range jobs {
				select {
				case <-ctx.Done():
					results <- result{parser: parser.Name(), err: ctx.Err()}
					return
				default:
					log.Printf("Worker %d: Starting %s", workerID, parser.Name())
					err := ss.scrapeSource(ctx, parser)
					results <- result{parser: parser.Name(), err: err}
					if err != nil {
						log.Printf("Worker %d: Error scraping %s: %v", workerID, parser.Name(), err)
					} else {
						log.Printf("Worker %d: Completed %s", workerID, parser.Name())
					}
				}
			}
		}(w)
	}

	// Send jobs to workers
	go func() {
		for _, parser := range ss.parsers {
			jobs <- parser
		}
		close(jobs)
	}()

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var errors []error
	for res := range results {
		if res.err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", res.parser, res.err))
		}
	}

	log.Println("Concurrent scraping process completed")

	if len(errors) > 0 {
		log.Printf("Completed with %d errors", len(errors))
		return fmt.Errorf("scraping had %d errors", len(errors))
	}

	return nil
}

func (ss *ScraperService) scrapeSource(ctx context.Context, parser parsers.Parser) error {
	startTime := time.Now()
	log.Printf("Scraping %s...", parser.Name())

	var savedCount, errorCount int64

	properties, err := parser.Parse(ctx)
	if err != nil {
		errorCount++
		ss.metricsService.RecordParserRun(parser.Name(), 0, 0, errorCount, time.Since(startTime))
		return fmt.Errorf("failed to parse %s: %w", parser.Name(), err)
	}

	log.Printf("Found %d properties from %s", len(properties), parser.Name())

	// Batch save properties (much faster than one-by-one)
	if len(properties) > 0 {
		saved, errors := ss.batchSaveProperties(properties)
		savedCount = int64(saved)
		errorCount = int64(errors)

		// Calculate factors for saved properties (async to not block)
		go ss.calculateFactorsAsync(properties)
	}

	// Record metrics
	ss.metricsService.RecordParserRun(parser.Name(), int64(len(properties)), savedCount, errorCount, time.Since(startTime))

	return nil
}

// batchSaveProperties saves properties in batches for better performance
func (ss *ScraperService) batchSaveProperties(properties []models.Property) (saved int, errors int) {
	const batchSize = 100

	for i := 0; i < len(properties); i += batchSize {
		end := i + batchSize
		if end > len(properties) {
			end = len(properties)
		}

		batch := properties[i:end]

		// Check for duplicates and update existing
		for j := range batch {
			var existing models.Property
			result := database.DB.Where("source = ? AND external_id = ?", batch[j].Source, batch[j].ExternalID).First(&existing)

			if result.Error == nil {
				// Update existing
				batch[j].ID = existing.ID
				batch[j].CreatedAt = existing.CreatedAt
				batch[j].UpdatedAt = time.Now()
			}
		}

		// Batch insert/update
		if err := database.DB.Save(&batch).Error; err != nil {
			log.Printf("Error batch saving properties: %v", err)
			errors += len(batch)
		} else {
			saved += len(batch)
		}
	}

	return saved, errors
}

// calculateFactorsAsync calculates factors for properties asynchronously
func (ss *ScraperService) calculateFactorsAsync(properties []models.Property) {
	for i := range properties {
		if properties[i].ID > 0 {
			factors, err := ss.factorsService.CalculateFactors(&properties[i])
			if err != nil {
				continue
			}

			if err := ss.factorsService.SaveFactors(factors); err != nil {
				continue
			}
		}
	}
}

func (ss *ScraperService) saveProperty(property *models.Property) error {
	// Check if property with this ExternalID and Source already exists
	var existing models.Property
	result := database.DB.Where("source = ? AND external_id = ?", property.Source, property.ExternalID).First(&existing)

	if result.Error == nil {
		// Update existing
		property.ID = existing.ID
		property.CreatedAt = existing.CreatedAt
		property.UpdatedAt = time.Now()
		return database.DB.Save(property).Error
	}

	// Create new
	return database.DB.Create(property).Error
}
