package parsers

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
	
	"github.com/PuerkitoBio/goquery"
	"pricemap-go/models"
	"pricemap-go/utils"
)

// ZillowParser parses real estate data from zillow.com (USA)
type ZillowParser struct {
	*BaseParser
	geocoding *utils.GeocodingService
}

func NewZillowParser() *ZillowParser {
	return &ZillowParser{
		BaseParser: NewBaseParser("https://www.zillow.com"),
		geocoding:  utils.NewGeocodingService(),
	}
}

func (zp *ZillowParser) Name() string {
	return "zillow"
}

func (zp *ZillowParser) Parse(ctx context.Context) ([]models.Property, error) {
	var allProperties []models.Property
	
	// USA cities to parse
	cities := []string{
		"New York, NY", "Los Angeles, CA", "Chicago, IL", "Houston, TX", "Phoenix, AZ",
		"Philadelphia, PA", "San Antonio, TX", "San Diego, CA", "Dallas, TX", "San Jose, CA",
		"Austin, TX", "Jacksonville, FL", "Fort Worth, TX", "Columbus, OH", "Charlotte, NC",
		"San Francisco, CA", "Indianapolis, IN", "Seattle, WA", "Denver, CO", "Washington, DC",
		"Boston, MA", "El Paso, TX", "Nashville, TN", "Detroit, MI", "Oklahoma City, OK",
		"Portland, OR", "Las Vegas, NV", "Memphis, TN", "Louisville, KY", "Baltimore, MD",
	}
	
	// Parse for both sale and rent
	dealTypes := []struct {
		path string
		name string
	}{
		{"homes", "sale"},
		{"apartments", "rent"},
	}
	
	for _, city := range cities {
		for _, dealType := range dealTypes {
			properties, err := zp.parseCity(ctx, city, dealType.path, dealType.name)
			if err != nil {
				log.Printf("Error parsing %s/%s from Zillow: %v", city, dealType.name, err)
				continue
			}
			allProperties = append(allProperties, properties...)
			
			// Rate limiting
			time.Sleep(3 * time.Second)
		}
	}
	
	log.Printf("Parsed %d properties from Zillow", len(allProperties))
	return allProperties, nil
}

func (zp *ZillowParser) parseCity(ctx context.Context, city, path, dealType string) ([]models.Property, error) {
	var properties []models.Property
	
	// Zillow search URL
	url := fmt.Sprintf("%s/%s/%s/", zp.baseURL, path, strings.ReplaceAll(city, " ", "-"))
	
	body, err := zp.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Zillow listings: %w", err)
	}
	defer body.Close()
	
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}
	
	// Zillow property card selectors
	doc.Find("[data-test='property-card']").Each(func(i int, s *goquery.Selection) {
		property := zp.parseProperty(s, city)
		if property != nil {
			properties = append(properties, *property)
		}
	})
	
	// Alternative selector
	if len(properties) == 0 {
		doc.Find(".list-card").Each(func(i int, s *goquery.Selection) {
			property := zp.parseProperty(s, city)
			if property != nil {
				properties = append(properties, *property)
			}
		})
	}
	
	return properties, nil
}

func (zp *ZillowParser) parseProperty(s *goquery.Selection, city string) *models.Property {
	property := &models.Property{
		Source:    zp.Name(),
		ScrapedAt: time.Now(),
		IsActive:  true,
		Currency:  "USD",
		Country:   "United States",
		City:      strings.Split(city, ",")[0], // Extract city name
		Type:      "apartment", // Default
	}
	
	// Extract price
	priceText := strings.TrimSpace(s.Find("[data-test='property-card-price']").Text())
	if priceText == "" {
		priceText = strings.TrimSpace(s.Find(".list-card-price").Text())
	}
	
	price := zp.extractPrice(priceText)
	if price <= 0 {
		return nil
	}
	property.Price = price
	
	// Extract address
	address := strings.TrimSpace(s.Find("[data-test='property-card-addr']").Text())
	if address == "" {
		address = strings.TrimSpace(s.Find(".list-card-addr").Text())
	}
	property.Address = address
	
	// Extract URL and ID
	if href, exists := s.Find("a").Attr("href"); exists {
		if strings.HasPrefix(href, "http") {
			property.URL = href
		} else {
			property.URL = zp.baseURL + href
		}
		property.ExternalID = zp.extractIDFromURL(href)
	}
	
	// Extract property details
	detailsText := s.Find("[data-test='property-card-details']").Text()
	if detailsText == "" {
		detailsText = s.Find(".list-card-details").Text()
	}
	
	property.Area = zp.extractArea(detailsText)
	property.Bedrooms = zp.extractBedrooms(detailsText)
	property.Bathrooms = zp.extractBathrooms(detailsText)
	
	// Extract coordinates if available
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
	
	// Geocode if coordinates missing
	if property.Latitude == 0 && property.Longitude == 0 && property.Address != "" {
		lat, lng, err := zp.geocoding.GeocodeAddress(property.Address + ", " + city)
		if err == nil {
			property.Latitude = lat
			property.Longitude = lng
		}
	}
	
	return property
}

func (zp *ZillowParser) extractPrice(text string) float64 {
	// Remove currency symbols and formatting
	re := regexp.MustCompile(`[^\d.]`)
	cleaned := re.ReplaceAllString(text, "")
	
	price, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0
	}
	
	return price
}

func (zp *ZillowParser) extractArea(text string) float64 {
	// Look for patterns like "1,234 sq ft" - convert to sq m
	re := regexp.MustCompile(`(\d+(?:,\d+)?)\s*(?:sq\s*ft|sqft)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		areaStr := strings.ReplaceAll(matches[1], ",", "")
		if area, err := strconv.ParseFloat(areaStr, 64); err == nil {
			return area * 0.092903 // Convert sq ft to sq m
		}
	}
	return 0
}

func (zp *ZillowParser) extractBedrooms(text string) int {
	re := regexp.MustCompile(`(\d+)\s*(?:bed|bedroom|br|beds)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		if bedrooms, err := strconv.Atoi(matches[1]); err == nil {
			return bedrooms
		}
	}
	return 0
}

func (zp *ZillowParser) extractBathrooms(text string) int {
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(?:bath|bathroom|ba)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		if bathrooms, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return int(bathrooms)
		}
	}
	return 0
}

func (zp *ZillowParser) extractIDFromURL(url string) string {
	// Extract ID from Zillow URL
	re := regexp.MustCompile(`/(\d+)_zpid/`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	return url
}

