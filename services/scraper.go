package services

import (
	"context"
	"fmt"
	"log"
	"pricemap-go/database"
	"pricemap-go/models"
	"pricemap-go/parsers"
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
			// Add all parsers here
			parsers.NewCianParser(),      // Russia - 30+ cities, sale & rent
			parsers.NewRightmoveParser(), // UK - 25+ cities, sale & rent
			parsers.NewZillowParser(),    // USA - 30+ cities, sale & rent
			parsers.NewIdealistaParser(), // Spain - 20+ cities, sale & rent
			// Add more country-specific parsers as they are implemented
			// parsers.NewImmobilienScoutParser(), // Germany
			// parsers.NewLeboncoinParser(),      // France
			// parsers.NewRealtorCaParser(),      // Canada
		},
		factorsService: NewFactorsService(),
		metricsService: NewMetricsService(),
		cacheService:   NewCacheService(1 * time.Hour), // 1 hour TTL
	}
}

// ScrapeAll starts parsing all sources
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
	
	// Save each property
	for i := range properties {
		if err := ss.saveProperty(&properties[i]); err != nil {
			log.Printf("Error saving property: %v", err)
			errorCount++
			continue
		}
		
		savedCount++
		
		// Calculate and save factors
		if properties[i].ID > 0 {
			factors, err := ss.factorsService.CalculateFactors(&properties[i])
			if err != nil {
				log.Printf("Error calculating factors: %v", err)
				continue
			}
			
			if err := ss.factorsService.SaveFactors(factors); err != nil {
				log.Printf("Error saving factors: %v", err)
			}
		}
	}
	
	// Record metrics
	ss.metricsService.RecordParserRun(parser.Name(), int64(len(properties)), savedCount, errorCount, time.Since(startTime))
	
	return nil
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

