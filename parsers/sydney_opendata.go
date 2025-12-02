package parsers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"pricemap-go/models"
	"pricemap-go/utils"
)

// SydneyOpenDataParser parses Sydney Open Data for property information
// API: https://data.nsw.gov.au/
type SydneyOpenDataParser struct {
	*BaseParser
	geocoding *utils.GeocodingService
}

func NewSydneyOpenDataParser() *SydneyOpenDataParser {
	return &SydneyOpenDataParser{
		BaseParser: NewBaseParser("https://data.nsw.gov.au"),
		geocoding:  utils.NewGeocodingService(),
	}
}

func (syd *SydneyOpenDataParser) Name() string {
	return "sydney_opendata"
}

func (syd *SydneyOpenDataParser) Parse(ctx context.Context) ([]models.Property, error) {
	var properties []models.Property

	// NSW Open Data - Property Sales
	apiURL := "https://data.nsw.gov.au/api/3/action/datastore_search?resource_id=property_sales&limit=5000"

	body, err := syd.Fetch(ctx, apiURL)
	if err != nil {
		log.Printf("Error fetching Sydney Open Data: %v", err)
		return properties, nil
	}
	defer body.Close()

	var response struct {
		Success bool `json:"success"`
		Result  struct {
			Records []struct {
				Suburb  string  `json:"suburb"`
				Price   float64 `json:"price"`
				Bedrooms int    `json:"bedrooms"`
				Bathrooms float64 `json:"bathrooms"`
				Area     float64 `json:"area"`
				Address  string  `json:"address"`
				Lat      float64 `json:"latitude"`
				Lng      float64 `json:"longitude"`
			} `json:"records"`
		} `json:"result"`
	}

	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&response); err != nil {
		return properties, nil
	}

	for _, item := range response.Result.Records {
		if item.Price <= 0 {
			continue
		}

		property := &models.Property{
			Source:    syd.Name(),
			ExternalID: fmt.Sprintf("sydney_%s_%s", item.Suburb, item.Address),
			Country:   "Australia",
			City:      "Sydney",
			District:  item.Suburb,
			Address:   item.Address,
			Price:     item.Price,
			Currency:  "AUD",
			Type:      "apartment",
			Area:      item.Area,
			Bedrooms:  item.Bedrooms,
			Bathrooms: int(item.Bathrooms),
			ScrapedAt: time.Now(),
			IsActive:  true,
		}

		if item.Lat != 0 && item.Lng != 0 {
			property.Latitude = item.Lat
			property.Longitude = item.Lng
		} else if item.Address != "" {
			address := item.Address + ", " + item.Suburb + ", Sydney, Australia"
			lat, lng, err := syd.geocoding.GeocodeAddress(address)
			if err == nil {
				property.Latitude = lat
				property.Longitude = lng
			}
			time.Sleep(1 * time.Second)
		}

		properties = append(properties, *property)
	}

	log.Printf("Parsed %d properties from Sydney Open Data", len(properties))
	return properties, nil
}

