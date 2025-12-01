package main

import (
	"log"
	"time"

	"pricemap-go/config"
	"pricemap-go/database"
	"pricemap-go/models"
	"pricemap-go/utils"
)

func main() {
	// Load configuration
	config.Load()

	// Connect to database
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Get geocoding service
	geocoding := utils.NewGeocodingService()

	// Find properties without coordinates
	var properties []models.Property
	result := database.DB.Where("(latitude = 0 AND longitude = 0) OR latitude IS NULL OR longitude IS NULL").Find(&properties)
	if result.Error != nil {
		log.Fatalf("Failed to query properties: %v", result.Error)
	}

	log.Printf("Found %d properties without coordinates", len(properties))

	// Geocode each property
	geocoded := 0
	failed := 0

	for i, property := range properties {
		// Build address from available data
		address := property.Address
		if address == "" {
			// Use city and country if address is missing
			if property.City != "" {
				address = property.City
				if property.Country != "" {
					address += ", " + property.Country
				}
			} else if property.Country != "" {
				address = property.Country
			} else {
				log.Printf("Skipping property %d: no address or location data", property.ID)
				failed++
				continue
			}
		}

		log.Printf("[%d/%d] Geocoding: %s", i+1, len(properties), address)

		// Geocode address
		lat, lng, err := geocoding.GeocodeAddress(address)
		if err != nil {
			log.Printf("Failed to geocode %s: %v", address, err)
			failed++
			// Rate limiting - wait before next request
			time.Sleep(2 * time.Second)
			continue
		}

		// Update property
		property.Latitude = lat
		property.Longitude = lng
		if result := database.DB.Save(&property); result.Error != nil {
			log.Printf("Failed to save property %d: %v", property.ID, result.Error)
			failed++
			continue
		}

		geocoded++
		log.Printf("✓ Geocoded: %s -> (%.6f, %.6f)", address, lat, lng)

		// Rate limiting for Nominatim (1 request per second)
		time.Sleep(1 * time.Second)
	}

	log.Printf("\nGeocoding completed:")
	log.Printf("  Successfully geocoded: %d", geocoded)
	log.Printf("  Failed: %d", failed)
	log.Printf("  Total: %d", len(properties))
}

