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

// RightmoveParser parses real estate data from rightmove.co.uk (UK)
type RightmoveParser struct {
	*BaseParser
	geocoding *utils.GeocodingService
}

func NewRightmoveParser() *RightmoveParser {
	return &RightmoveParser{
		BaseParser: NewBaseParser("https://www.rightmove.co.uk"),
		geocoding:  utils.NewGeocodingService(),
	}
}

func (rp *RightmoveParser) Name() string {
	return "rightmove"
}

func (rp *RightmoveParser) Parse(ctx context.Context) ([]models.Property, error) {
	var allProperties []models.Property
	
	// UK cities to parse
	cities := []string{
		"London", "Manchester", "Birmingham", "Liverpool", "Leeds",
		"Glasgow", "Edinburgh", "Bristol", "Cardiff", "Belfast",
		"Newcastle", "Sheffield", "Leicester", "Coventry", "Nottingham",
		"Southampton", "Portsmouth", "Brighton", "Reading", "Oxford",
		"Cambridge", "York", "Bath", "Norwich", "Exeter",
	}
	
	// Parse for both sale and rent
	dealTypes := []struct {
		path string
		name string
	}{
		{"property-for-sale", "sale"},
		{"property-to-rent", "rent"},
	}
	
	for _, city := range cities {
		for _, dealType := range dealTypes {
			properties, err := rp.parseCity(ctx, city, dealType.path, dealType.name)
			if err != nil {
				log.Printf("Error parsing %s/%s from Rightmove: %v", city, dealType.name, err)
				continue
			}
			allProperties = append(allProperties, properties...)
			
			// Rate limiting
			time.Sleep(2 * time.Second)
		}
	}
	
	log.Printf("Parsed %d properties from Rightmove", len(allProperties))
	return allProperties, nil
}

func (rp *RightmoveParser) parseCity(ctx context.Context, city, path string, _ string) ([]models.Property, error) {
	var properties []models.Property
	
	// Rightmove search URL - using location search
	// Note: Location identifiers need to be mapped for each city
	url := fmt.Sprintf("%s/%s/find.html?locationIdentifier=&minBedrooms=&maxBedrooms=&minPrice=&maxPrice=&propertyTypes=&mustHave=&dontShow=&furnishTypes=&keywords=%s",
		rp.baseURL,
		path,
		city,
	)
	
	body, err := rp.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Rightmove listings: %w", err)
	}
	defer body.Close()
	
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}
	
	// Rightmove property card selector
	doc.Find(".l-searchResults .propertyCard").Each(func(i int, s *goquery.Selection) {
		property := rp.parseProperty(s)
		if property != nil {
			property.City = city
			properties = append(properties, *property)
		}
	})
	
	return properties, nil
}

func (rp *RightmoveParser) parseProperty(s *goquery.Selection) *models.Property {
	property := &models.Property{
		Source:    rp.Name(),
		ScrapedAt: time.Now(),
		IsActive:  true,
		Currency:  "GBP",
		Country:   "United Kingdom",
		Type:      "apartment", // Default
	}
	
	// Extract price
	priceText := strings.TrimSpace(s.Find(".propertyCard-price").Text())
	if priceText == "" {
		priceText = strings.TrimSpace(s.Find("[data-test='property-price']").Text())
	}
	
	price := rp.extractPrice(priceText)
	if price <= 0 {
		return nil
	}
	property.Price = price
	
	// Extract address
	address := strings.TrimSpace(s.Find(".propertyCard-address").Text())
	if address == "" {
		address = strings.TrimSpace(s.Find("[data-test='property-address']").Text())
	}
	property.Address = address
	
	// Extract city from address
	if address != "" {
		parts := strings.Split(address, ",")
		if len(parts) > 0 {
			property.City = strings.TrimSpace(parts[len(parts)-1])
		}
	}
	
	// Extract URL and ID
	if href, exists := s.Find("a.propertyCard-link").Attr("href"); exists {
		if strings.HasPrefix(href, "http") {
			property.URL = href
		} else {
			property.URL = rp.baseURL + href
		}
		property.ExternalID = rp.extractIDFromURL(href)
	}
	
	// Extract property details
	detailsText := s.Find(".propertyCard-details").Text()
	property.Area = rp.extractArea(detailsText)
	property.Bedrooms = rp.extractBedrooms(detailsText)
	
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
		lat, lng, err := rp.geocoding.GeocodeAddress(property.Address + ", UK")
		if err == nil {
			property.Latitude = lat
			property.Longitude = lng
		}
	}
	
	return property
}

func (rp *RightmoveParser) extractPrice(text string) float64 {
	// Remove currency symbols and formatting
	re := regexp.MustCompile(`[^\d.]`)
	cleaned := re.ReplaceAllString(text, "")
	
	price, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0
	}
	
	// Rightmove prices are usually in full (not thousands)
	return price
}

func (rp *RightmoveParser) extractArea(text string) float64 {
	// Look for patterns like "1,234 sq ft" or "123 sq m"
	re := regexp.MustCompile(`(\d+(?:,\d+)?)\s*(?:sq\s*ft|sq\s*m|sqft|sqm)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		areaStr := strings.ReplaceAll(matches[1], ",", "")
		if area, err := strconv.ParseFloat(areaStr, 64); err == nil {
			// Convert sq ft to sq m if needed
			if strings.Contains(strings.ToLower(text), "sq ft") || strings.Contains(strings.ToLower(text), "sqft") {
				area = area * 0.092903 // Convert sq ft to sq m
			}
			return area
		}
	}
	return 0
}

func (rp *RightmoveParser) extractBedrooms(text string) int {
	// Look for patterns like "3 bed" or "3 bedrooms"
	re := regexp.MustCompile(`(\d+)\s*(?:bed|bedroom)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		if bedrooms, err := strconv.Atoi(matches[1]); err == nil {
			return bedrooms
		}
	}
	return 0
}

func (rp *RightmoveParser) extractIDFromURL(url string) string {
	// Extract ID from URL like /properties/12345678
	re := regexp.MustCompile(`/properties/(\d+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	return url
}

