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

// ParisOpenDataParser parses Paris Open Data for property information
// API: https://opendata.paris.fr/
type ParisOpenDataParser struct {
	*BaseParser
	geocoding *utils.GeocodingService
}

func NewParisOpenDataParser() *ParisOpenDataParser {
	return &ParisOpenDataParser{
		BaseParser: NewBaseParser("https://opendata.paris.fr"),
		geocoding:  utils.NewGeocodingService(),
	}
}

func (par *ParisOpenDataParser) Name() string {
	return "paris_opendata"
}

func (par *ParisOpenDataParser) Parse(ctx context.Context) ([]models.Property, error) {
	var properties []models.Property

	// Paris Open Data - Real Estate Transactions
	apiURL := "https://opendata.paris.fr/api/records/1.0/search/?dataset=logements-encadrement-des-loyers&rows=5000"

	body, err := par.Fetch(ctx, apiURL)
	if err != nil {
		log.Printf("Error fetching Paris Open Data: %v", err)
		return properties, nil
	}
	defer body.Close()

	var response struct {
		Records []struct {
			Fields struct {
				Arrondissement string    `json:"arrondissement"`
				Loyer          float64   `json:"loyer_m2"`
				Surface        float64   `json:"surface"`
				Address        string    `json:"adresse"`
				Coords         []float64 `json:"geo_point_2d"`
			} `json:"fields"`
		} `json:"records"`
	}

	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&response); err != nil {
		return properties, nil
	}

	for _, record := range response.Records {
		fields := record.Fields
		if fields.Loyer <= 0 || fields.Surface <= 0 {
			continue
		}

		// Calculate annual rent (monthly * 12)
		annualRent := fields.Loyer * fields.Surface * 12

		property := &models.Property{
			Source:     par.Name(),
			ExternalID: fmt.Sprintf("paris_%s_%s", fields.Arrondissement, fields.Address),
			Country:    "France",
			City:       "Paris",
			District:   fields.Arrondissement,
			Address:    fields.Address,
			Price:      annualRent,
			Currency:   "EUR",
			Type:       "apartment",
			Area:       fields.Surface,
			ScrapedAt:  time.Now(),
			IsActive:   true,
		}

		if len(fields.Coords) >= 2 {
			property.Latitude = fields.Coords[0]
			property.Longitude = fields.Coords[1]
		} else if fields.Address != "" {
			address := fields.Address + ", Paris, France"
			lat, lng, err := par.geocoding.GeocodeAddress(address)
			if err == nil {
				property.Latitude = lat
				property.Longitude = lng
			}
			time.Sleep(1 * time.Second)
		}

		properties = append(properties, *property)
	}

	log.Printf("Parsed %d properties from Paris Open Data", len(properties))
	return properties, nil
}
