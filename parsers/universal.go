package parsers

import (
	"context"
	"log"
	"pricemap-go/models"
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

