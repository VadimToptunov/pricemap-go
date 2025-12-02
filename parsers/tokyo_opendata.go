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

// TokyoOpenDataParser parses Tokyo Open Data for property information
// API: https://portal.data.metro.tokyo.lg.jp/
type TokyoOpenDataParser struct {
	*BaseParser
	geocoding *utils.GeocodingService
}

func NewTokyoOpenDataParser() *TokyoOpenDataParser {
	return &TokyoOpenDataParser{
		BaseParser: NewBaseParser("https://portal.data.metro.tokyo.lg.jp"),
		geocoding:  utils.NewGeocodingService(),
	}
}

func (tok *TokyoOpenDataParser) Name() string {
	return "tokyo_opendata"
}

func (tok *TokyoOpenDataParser) Parse(ctx context.Context) ([]models.Property, error) {
	var properties []models.Property

	// Tokyo Open Data - Real Estate Prices
	// Using CKAN API
	apiURL := "https://portal.data.metro.tokyo.lg.jp/api/3/action/datastore_search?resource_id=real_estate_prices&limit=5000"

	body, err := tok.Fetch(ctx, apiURL)
	if err != nil {
		log.Printf("Error fetching Tokyo Open Data: %v", err)
		return properties, nil
	}
	defer body.Close()

	var response struct {
		Success bool `json:"success"`
		Result  struct {
			Records []struct {
				Ward     string  `json:"ward"`
				Price    float64 `json:"price"`
				Area     float64 `json:"area"`
				Rooms    int     `json:"rooms"`
				Address  string  `json:"address"`
				Lat      float64 `json:"lat"`
				Lng      float64 `json:"lng"`
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
			Source:    tok.Name(),
			ExternalID: fmt.Sprintf("tokyo_%s_%s", item.Ward, item.Address),
			Country:   "Japan",
			City:      "Tokyo",
			District:  item.Ward,
			Address:   item.Address,
			Price:     item.Price,
			Currency:  "JPY",
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
			address := item.Address + ", Tokyo, Japan"
			lat, lng, err := tok.geocoding.GeocodeAddress(address)
			if err == nil {
				property.Latitude = lat
				property.Longitude = lng
			}
			time.Sleep(1 * time.Second)
		}

		properties = append(properties, *property)
	}

	log.Printf("Parsed %d properties from Tokyo Open Data", len(properties))
	return properties, nil
}

