package parsers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	
	"github.com/PuerkitoBio/goquery"
	"pricemap-go/models"
)

// ExampleParser - example parser (can be adapted for specific sites)
type ExampleParser struct {
	*BaseParser
}

func NewExampleParser() *ExampleParser {
	return &ExampleParser{
		BaseParser: NewBaseParser("https://example-real-estate.com"),
	}
}

func (ep *ExampleParser) Name() string {
	return "example_parser"
}

func (ep *ExampleParser) Parse(ctx context.Context) ([]models.Property, error) {
	var properties []models.Property
	
	// Example: parsing property listings
	url := fmt.Sprintf("%s/listings", ep.baseURL)
	
	body, err := ep.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch listings: %w", err)
	}
	defer body.Close()
	
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}
	
	doc.Find(".property-listing").Each(func(i int, s *goquery.Selection) {
		property := ep.parseProperty(s)
		if property != nil {
			properties = append(properties, *property)
		}
	})
	
	log.Printf("Parsed %d properties from %s", len(properties), ep.Name())
	return properties, nil
}

func (ep *ExampleParser) parseProperty(s *goquery.Selection) *models.Property {
	property := &models.Property{
		Source:    ep.Name(),
		ScrapedAt: time.Now(),
		IsActive:  true,
		Currency:  "USD",
	}
	
	// Parse price
	priceText := strings.TrimSpace(s.Find(".price").Text())
	priceText = strings.ReplaceAll(priceText, ",", "")
	priceText = strings.ReplaceAll(priceText, "$", "")
	if price, err := strconv.ParseFloat(priceText, 64); err == nil {
		property.Price = price
	} else {
		return nil
	}
	
	// Parse address
	property.Address = strings.TrimSpace(s.Find(".address").Text())
	
	// Parse type
	property.Type = strings.ToLower(strings.TrimSpace(s.Find(".type").Text()))
	
	// Parse area
	areaText := strings.TrimSpace(s.Find(".area").Text())
	if area, err := strconv.ParseFloat(areaText, 64); err == nil {
		property.Area = area
	}
	
	// Parse rooms
	roomsText := strings.TrimSpace(s.Find(".rooms").Text())
	if rooms, err := strconv.Atoi(roomsText); err == nil {
		property.Rooms = rooms
	}
	
	// Parse URL
	if href, exists := s.Find("a").Attr("href"); exists {
		if strings.HasPrefix(href, "http") {
			property.URL = href
		} else {
			property.URL = ep.baseURL + href
		}
		property.ExternalID = extractIDFromURL(href)
	}
	
	// Parse coordinates (if available)
	if lat, exists := s.Attr("data-lat"); exists {
		if latitude, err := strconv.ParseFloat(lat, 64); err == nil {
			property.Latitude = latitude
		}
	}
	if lng, exists := s.Attr("data-lng"); exists {
		if longitude, err := strconv.ParseFloat(lng, 64); err == nil {
			property.Longitude = longitude
		}
	}
	
	return property
}

func extractIDFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return url
}

