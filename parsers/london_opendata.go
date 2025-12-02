package parsers

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"pricemap-go/models"
	"pricemap-go/utils"
)

// LondonOpenDataParser parses London Data Store for property prices
// API: https://data.london.gov.uk/dataset/uk-house-price-index
type LondonOpenDataParser struct {
	*BaseParser
	geocoding *utils.GeocodingService
}

func NewLondonOpenDataParser() *LondonOpenDataParser {
	return &LondonOpenDataParser{
		BaseParser: NewBaseParser("https://data.london.gov.uk"),
		geocoding:  utils.NewGeocodingService(),
	}
}

func (ldn *LondonOpenDataParser) Name() string {
	return "london_opendata"
}

func (ldn *LondonOpenDataParser) Parse(ctx context.Context) ([]models.Property, error) {
	var properties []models.Property

	// London Data Store - UK House Price Index
	// Using CSV format
	apiURL := "https://data.london.gov.uk/download/uk-house-price-index/70c07674-14bb-4285-989a-c888dff80102/house-price-index-2023.csv"

	body, err := ldn.Fetch(ctx, apiURL)
	if err != nil {
		log.Printf("Error fetching London Open Data: %v", err)
		return nil, err
	}
	defer body.Close()

	reader := csv.NewReader(body)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	// Skip header
	for i, record := range records {
		if i == 0 {
			continue
		}

		if len(record) < 5 {
			continue
		}

		// Parse price (assuming format: area, date, price, ...)
		priceStr := strings.ReplaceAll(record[2], ",", "")
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil || price <= 0 {
			continue
		}

		area := record[0]
		if area == "" {
			continue
		}

		property := &models.Property{
			Source:    ldn.Name(),
			ExternalID: fmt.Sprintf("london_%s_%d", area, i),
			Country:   "United Kingdom",
			City:      "London",
			District:  area,
			Price:     price,
			Currency:  "GBP",
			Type:      "apartment",
			ScrapedAt: time.Now(),
			IsActive:  true,
		}

		// Geocode area
		address := area + ", London, UK"
		lat, lng, err := ldn.geocoding.GeocodeAddress(address)
		if err == nil {
			property.Latitude = lat
			property.Longitude = lng
		}
		time.Sleep(1 * time.Second) // Rate limiting

		properties = append(properties, *property)
	}

	log.Printf("Parsed %d properties from London Open Data", len(properties))
	return properties, nil
}

