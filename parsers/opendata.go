package parsers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"
	
	"pricemap-go/models"
	"pricemap-go/utils"
)

// OpenDataParser parses data from government open data portals
// This is a generic parser that can be configured for different portals
type OpenDataParser struct {
	*BaseParser
	geocoding    *utils.GeocodingService
	apiEndpoint  string
	dataFormat   string // "json", "csv", "xml"
	country      string
	city         string
}

// NewOpenDataParser creates a new open data parser
func NewOpenDataParser(apiEndpoint, country, city string) *OpenDataParser {
	return &OpenDataParser{
		BaseParser: NewBaseParser(apiEndpoint),
		geocoding:  utils.NewGeocodingService(),
		apiEndpoint: apiEndpoint,
		dataFormat:  "json",
		country:     country,
		city:        city,
	}
}

func (odp *OpenDataParser) Name() string {
	return fmt.Sprintf("opendata_%s_%s", odp.country, odp.city)
}

func (odp *OpenDataParser) Parse(ctx context.Context) ([]models.Property, error) {
	var properties []models.Property
	
	// Fetch data from API
	body, err := odp.Fetch(ctx, odp.apiEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch open data: %w", err)
	}
	defer body.Close()
	
	// Parse based on format
	switch odp.dataFormat {
	case "json":
		properties, err = odp.parseJSON(body)
	case "csv":
		properties, err = odp.parseCSV(body)
	default:
		return nil, fmt.Errorf("unsupported data format: %s", odp.dataFormat)
	}
	
	if err != nil {
		return nil, err
	}
	
	log.Printf("Parsed %d properties from %s", len(properties), odp.Name())
	return properties, nil
}

func (odp *OpenDataParser) parseJSON(body io.ReadCloser) ([]models.Property, error) {
	var properties []models.Property
	
	// Generic JSON structure - adapt based on actual API response
	var data struct {
		Results []struct {
			Address   string  `json:"address"`
			Price     float64 `json:"price"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
			Area      float64 `json:"area"`
			Rooms     int     `json:"rooms"`
			Type      string  `json:"type"`
			ID        string  `json:"id"`
		} `json:"results"`
	}
	
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}
	
	for _, item := range data.Results {
		property := &models.Property{
			Source:     odp.Name(),
			ExternalID: item.ID,
			Address:    item.Address,
			Price:      item.Price,
			Latitude:   item.Latitude,
			Longitude:  item.Longitude,
			Area:       item.Area,
			Rooms:      item.Rooms,
			Type:       odp.normalizeType(item.Type),
			Country:    odp.country,
			City:       odp.city,
			ScrapedAt:  time.Now(),
			IsActive:   true,
			Currency:   odp.getCurrency(),
		}
		
		// Geocode if coordinates missing
		if property.Latitude == 0 && property.Longitude == 0 && property.Address != "" {
			lat, lng, err := odp.geocoding.GeocodeAddress(property.Address + ", " + property.City)
			if err == nil {
				property.Latitude = lat
				property.Longitude = lng
			}
		}
		
		properties = append(properties, *property)
	}
	
	return properties, nil
}

func (odp *OpenDataParser) parseCSV(body io.ReadCloser) ([]models.Property, error) {
	// CSV parsing would go here
	// For now, return empty - can be implemented later
	return nil, fmt.Errorf("CSV parsing not yet implemented")
}

func (odp *OpenDataParser) normalizeType(propType string) string {
	propType = strings.ToLower(propType)
	switch {
	case strings.Contains(propType, "apartment") || strings.Contains(propType, "flat"):
		return "apartment"
	case strings.Contains(propType, "house"):
		return "house"
	case strings.Contains(propType, "room"):
		return "room"
	default:
		return "apartment"
	}
}

func (odp *OpenDataParser) getCurrency() string {
	switch odp.country {
	case "USA", "United States":
		return "USD"
	case "UK", "United Kingdom":
		return "GBP"
	case "Russia":
		return "RUB"
	case "Germany":
		return "EUR"
	case "France":
		return "EUR"
	case "Spain":
		return "EUR"
	default:
		return "USD"
	}
}

