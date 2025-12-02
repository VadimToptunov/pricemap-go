package parsers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"pricemap-go/models"
	"pricemap-go/utils"
)

// NYCOpenDataParser parses NYC Open Data for property sales
// API: https://data.cityofnewyork.us/Housing-Development/Property-Sales/22z6-9x9z
type NYCOpenDataParser struct {
	*BaseParser
	geocoding *utils.GeocodingService
}

func NewNYCOpenDataParser() *NYCOpenDataParser {
	return &NYCOpenDataParser{
		BaseParser: NewBaseParser("https://data.cityofnewyork.us"),
		geocoding:  utils.NewGeocodingService(),
	}
}

func (nyc *NYCOpenDataParser) Name() string {
	return "nyc_opendata"
}

func (nyc *NYCOpenDataParser) Parse(ctx context.Context) ([]models.Property, error) {
	var properties []models.Property

	// NYC Open Data API endpoint for property sales
	// Using Socrata API format
	apiURL := "https://data.cityofnewyork.us/resource/22z6-9x9z.json?$limit=5000&$where=sale_price>0"

	body, err := nyc.Fetch(ctx, apiURL)
	if err != nil {
		log.Printf("Error fetching NYC Open Data: %v", err)
		return nil, err
	}
	defer body.Close()

	var data []struct {
		Borough      string `json:"borough"`
		Neighborhood string `json:"neighborhood"`
		Address      string `json:"address"`
		SalePrice    string `json:"sale_price"`
		SaleDate     string `json:"sale_date"`
		Residential  string `json:"residential_units"`
		Commercial   string `json:"commercial_units"`
		TotalUnits   string `json:"total_units"`
		LandSquare   string `json:"land_square_feet"`
		GrossSquare  string `json:"gross_square_feet"`
		YearBuilt    string `json:"year_built"`
		Latitude     string `json:"latitude"`
		Longitude    string `json:"longitude"`
	}

	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode NYC data: %w", err)
	}

	for _, item := range data {
		price, _ := strconv.ParseFloat(item.SalePrice, 64)
		if price <= 0 {
			continue
		}

		property := &models.Property{
			Source:     nyc.Name(),
			ExternalID: fmt.Sprintf("nyc_%s_%s", item.Address, item.SaleDate),
			Country:    "United States",
			City:       "New York",
			District:   item.Neighborhood,
			Address:    item.Address,
			Price:      price,
			Currency:   "USD",
			Type:       "apartment",
			ScrapedAt:  time.Now(),
			IsActive:   true,
		}

		// Parse coordinates
		if item.Latitude != "" && item.Longitude != "" {
			if lat, err := strconv.ParseFloat(item.Latitude, 64); err == nil {
				property.Latitude = lat
			}
			if lng, err := strconv.ParseFloat(item.Longitude, 64); err == nil {
				property.Longitude = lng
			}
		}

		// Parse area (convert sq ft to sq m)
		if item.GrossSquare != "" {
			if area, err := strconv.ParseFloat(item.GrossSquare, 64); err == nil {
				property.Area = area * 0.092903 // sq ft to sq m
			}
		}

		// Parse year built
		if item.YearBuilt != "" {
			if year, err := strconv.Atoi(item.YearBuilt); err == nil {
				property.YearBuilt = year
			}
		}

		// Parse units as rooms
		if item.Residential != "" {
			if units, err := strconv.Atoi(item.Residential); err == nil {
				property.Rooms = units
			}
		}

		// Geocode if coordinates missing
		if property.Latitude == 0 && property.Longitude == 0 && property.Address != "" {
			address := property.Address + ", " + property.City + ", NY"
			lat, lng, err := nyc.geocoding.GeocodeAddress(address)
			if err == nil {
				property.Latitude = lat
				property.Longitude = lng
			}
			time.Sleep(1 * time.Second) // Rate limiting
		}

		properties = append(properties, *property)
	}

	log.Printf("Parsed %d properties from NYC Open Data", len(properties))
	return properties, nil
}
