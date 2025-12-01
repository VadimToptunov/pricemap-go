package parsers

import (
	"context"
	"log"
	"pricemap-go/models"
	"pricemap-go/utils"
	"time"
)

// UniversalParser is a meta-parser that uses multiple sources
// to cover all cities worldwide
type UniversalParser struct {
	*BaseParser
	parsers []Parser
}

func NewUniversalParser() *UniversalParser {
	// Initialize all available parsers
	parsers := []Parser{
		NewCianParser(),      // Russia
		NewRightmoveParser(), // UK
		// Add more country-specific parsers here
	}
	
	return &UniversalParser{
		BaseParser: NewBaseParser(""),
		parsers:    parsers,
	}
}

func (up *UniversalParser) Name() string {
	return "universal"
}

func (up *UniversalParser) GetBaseURL() string {
	return ""
}

func (up *UniversalParser) Parse(ctx context.Context) ([]models.Property, error) {
	var allProperties []models.Property
	
	// Run all parsers in parallel (with context cancellation support)
	for _, parser := range up.parsers {
		properties, err := parser.Parse(ctx)
		if err != nil {
			log.Printf("Error in parser %s: %v", parser.Name(), err)
			continue
		}
		allProperties = append(allProperties, properties...)
	}
	
	log.Printf("Universal parser collected %d properties from %d sources", 
		len(allProperties), len(up.parsers))
	
	return allProperties, nil
}

// CityBasedParser is a helper for parsers that need to iterate over cities
type CityBasedParser struct {
	cities []utils.City
}

// GetCitiesForCountry returns cities for a specific country
func GetCitiesForCountry(country string) []utils.City {
	return utils.GetCitiesByCountry(country)
}

// GetAllMajorCities returns all major cities worldwide
func GetAllMajorCities() []utils.City {
	return utils.GetMajorCities()
}

// ParseWithCities is a helper function for parsers that need to iterate over cities
func ParseWithCities(ctx context.Context, parser Parser, cities []utils.City, 
	parseFunc func(context.Context, utils.City) ([]models.Property, error)) ([]models.Property, error) {
	
	var allProperties []models.Property
	
	for _, city := range cities {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return allProperties, ctx.Err()
		default:
		}
		
		properties, err := parseFunc(ctx, city)
		if err != nil {
			log.Printf("Error parsing %s, %s: %v", city.Name, city.Country, err)
			continue
		}
		
		// Set city information for all properties
		for i := range properties {
			properties[i].City = city.Name
			properties[i].Country = city.Country
			if properties[i].Latitude == 0 && properties[i].Longitude == 0 {
				properties[i].Latitude = city.Latitude
				properties[i].Longitude = city.Longitude
			}
		}
		
		allProperties = append(allProperties, properties...)
		
		// Rate limiting between cities
		time.Sleep(1 * time.Second)
	}
	
	return allProperties, nil
}

