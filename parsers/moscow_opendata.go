package parsers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"pricemap-go/models"
	"pricemap-go/utils"
)

// MoscowOpenDataParser parses Moscow Open Data for property information
// API: https://data.mos.ru/
type MoscowOpenDataParser struct {
	*BaseParser
	geocoding *utils.GeocodingService
}

func NewMoscowOpenDataParser() *MoscowOpenDataParser {
	return &MoscowOpenDataParser{
		BaseParser: NewBaseParser("https://data.mos.ru"),
		geocoding:  utils.NewGeocodingService(),
	}
}

func (mos *MoscowOpenDataParser) Name() string {
	return "moscow_opendata"
}

func (mos *MoscowOpenDataParser) Parse(ctx context.Context) ([]models.Property, error) {
	var properties []models.Property

	// Moscow Open Data - Real Estate Transactions
	apiURL := "https://data.mos.ru/api/v1/datasets/real_estate_transactions/rows?$top=5000"

	body, err := mos.Fetch(ctx, apiURL)
	if err != nil {
		log.Printf("Error fetching Moscow Open Data: %v", err)
		return properties, nil
	}
	defer body.Close()

	var data []struct {
		Cells struct {
			District string  `json:"district"`
			Address  string  `json:"address"`
			Price    string  `json:"price"`
			Area     string  `json:"area"`
			Rooms    string  `json:"rooms"`
			Lat      string  `json:"latitude"`
			Lng      string  `json:"longitude"`
		} `json:"Cells"`
	}

	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&data); err != nil {
		return properties, nil
	}

	for _, item := range data {
		priceStr := strings.ReplaceAll(item.Cells.Price, " ", "")
		priceStr = strings.ReplaceAll(priceStr, ",", ".")
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil || price <= 0 {
			continue
		}

		property := &models.Property{
			Source:    mos.Name(),
			ExternalID: fmt.Sprintf("moscow_%s_%s", item.Cells.District, item.Cells.Address),
			Country:   "Russia",
			City:      "Moscow",
			District:  item.Cells.District,
			Address:   item.Cells.Address,
			Price:     price,
			Currency:  "RUB",
			Type:      "apartment",
			ScrapedAt: time.Now(),
			IsActive:  true,
		}

		// Parse area
		if item.Cells.Area != "" {
			areaStr := strings.ReplaceAll(item.Cells.Area, ",", ".")
			if area, err := strconv.ParseFloat(areaStr, 64); err == nil {
				property.Area = area
			}
		}

		// Parse rooms
		if item.Cells.Rooms != "" {
			if rooms, err := strconv.Atoi(item.Cells.Rooms); err == nil {
				property.Rooms = rooms
			}
		}

		// Parse coordinates
		if item.Cells.Lat != "" && item.Cells.Lng != "" {
			if lat, err := strconv.ParseFloat(item.Cells.Lat, 64); err == nil {
				property.Latitude = lat
			}
			if lng, err := strconv.ParseFloat(item.Cells.Lng, 64); err == nil {
				property.Longitude = lng
			}
		}

		// Geocode if coordinates missing
		if property.Latitude == 0 && property.Longitude == 0 && property.Address != "" {
			address := property.Address + ", Москва, Россия"
			lat, lng, err := mos.geocoding.GeocodeAddress(address)
			if err == nil {
				property.Latitude = lat
				property.Longitude = lng
			}
			time.Sleep(1 * time.Second)
		}

		properties = append(properties, *property)
	}

	log.Printf("Parsed %d properties from Moscow Open Data", len(properties))
	return properties, nil
}

