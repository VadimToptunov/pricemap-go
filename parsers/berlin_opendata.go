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

// BerlinOpenDataParser parses Berlin Open Data for property information
// API: https://daten.berlin.de/
type BerlinOpenDataParser struct {
	*BaseParser
	geocoding *utils.GeocodingService
}

func NewBerlinOpenDataParser() *BerlinOpenDataParser {
	return &BerlinOpenDataParser{
		BaseParser: NewBaseParser("https://daten.berlin.de"),
		geocoding:  utils.NewGeocodingService(),
	}
}

func (ber *BerlinOpenDataParser) Name() string {
	return "berlin_opendata"
}

func (ber *BerlinOpenDataParser) Parse(ctx context.Context) ([]models.Property, error) {
	var properties []models.Property

	// Berlin Open Data - Real Estate Prices
	// Using CKAN API format
	apiURL := "https://daten.berlin.de/api/3/action/datastore_search?resource_id=real_estate_prices&limit=5000"

	body, err := ber.Fetch(ctx, apiURL)
	if err != nil {
		log.Printf("Error fetching Berlin Open Data: %v", err)
		// Try alternative approach - return empty if API not available
		return properties, nil
	}
	defer body.Close()

	var response struct {
		Success bool `json:"success"`
		Result  struct {
			Records []struct {
				District string  `json:"district"`
				Price    float64 `json:"price_per_sqm"`
				Area     float64 `json:"area_sqm"`
				Rooms    int     `json:"rooms"`
				Address  string  `json:"address"`
				Lat      float64 `json:"latitude"`
				Lng      float64 `json:"longitude"`
			} `json:"records"`
		} `json:"result"`
	}

	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&response); err != nil {
		// If parsing fails, return empty (API might have different format)
		return properties, nil
	}

	for _, item := range response.Result.Records {
		if item.Price <= 0 {
			continue
		}

		totalPrice := item.Price * item.Area
		if totalPrice <= 0 {
			continue
		}

		property := &models.Property{
			Source:    ber.Name(),
			ExternalID: fmt.Sprintf("berlin_%s_%s", item.District, item.Address),
			Country:   "Germany",
			City:      "Berlin",
			District:  item.District,
			Address:   item.Address,
			Price:     totalPrice,
			Currency:  "EUR",
			Type:      "apartment",
			Area:      item.Area,
			Rooms:     item.Rooms,
			ScrapedAt: time.Now(),
			IsActive:  true,
		}

		if item.Lat != 0 && item.Lng != 0 {
			property.Latitude = item.Lat
			property.Longitude = item.Lng
		} else if item.Address != "" {
			address := item.Address + ", Berlin, Germany"
			lat, lng, err := ber.geocoding.GeocodeAddress(address)
			if err == nil {
				property.Latitude = lat
				property.Longitude = lng
			}
			time.Sleep(1 * time.Second)
		}

		properties = append(properties, *property)
	}

	log.Printf("Parsed %d properties from Berlin Open Data", len(properties))
	return properties, nil
}

