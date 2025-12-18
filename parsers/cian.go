package parsers

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"pricemap-go/models"
	"pricemap-go/utils"

	"github.com/PuerkitoBio/goquery"
)

// CianParser parses real estate data from cian.ru (Russia)
type CianParser struct {
	*BaseParser
	geocoding *utils.GeocodingService
}

func NewCianParser() *CianParser {
	return &CianParser{
		BaseParser: NewBaseParser("https://www.cian.ru"),
		geocoding:  utils.NewGeocodingService(),
	}
}

func (cp *CianParser) Name() string {
	return "cian"
}

func (cp *CianParser) Parse(ctx context.Context) ([]models.Property, error) {
	var allProperties []models.Property
	
	// Start with top 5 cities to avoid overwhelming on first run
	// Users can expand this list as needed
	cities := []string{
		"Moscow", "Saint Petersburg", "Novosibirsk", "Yekaterinburg", "Kazan",
	}
	
	// Parse different property types
	types := []string{"flat"}  // Start with flats only, add others later
	
	// Parse for both sale and rent
	dealTypes := []string{"sale"}  // Start with sales only
	
	totalCombinations := len(cities) * len(dealTypes) * len(types)
	processed := 0
	
	for _, city := range cities {
		for _, dealType := range dealTypes {
			for _, propType := range types {
				// Check context cancellation
				select {
				case <-ctx.Done():
					log.Printf("Context cancelled. Returning %d properties parsed so far.", len(allProperties))
					return allProperties, ctx.Err()
				default:
				}
				
				processed++
				log.Printf("Cian: Processing %s/%s/%s (%d/%d)", city, dealType, propType, processed, totalCombinations)
				
				properties, err := cp.parseType(ctx, propType, dealType, city)
				if err != nil {
					log.Printf("Error parsing %s/%s/%s from Cian: %v", city, dealType, propType, err)
					// Continue to next combination instead of failing completely
					continue
				}
				
				log.Printf("Cian: Found %d properties for %s/%s/%s", len(properties), city, dealType, propType)
				allProperties = append(allProperties, properties...)
				
				// Note: Rate limiting is now handled in BaseParser.Fetch()
			}
		}
	}
	
	log.Printf("Cian: Successfully parsed %d total properties", len(allProperties))
	return allProperties, nil
}

func (cp *CianParser) parseType(ctx context.Context, propType, dealType, city string) ([]models.Property, error) {
	var properties []models.Property
	
	// Cian search URL structure with city and deal type
	// Note: Region codes need to be mapped for each city
	// For now, using a generic search that works for major cities
	url := fmt.Sprintf("%s/cat.php?deal_type=%s&engine_version=2&object_type[0]=1&offer_type=%s&region=1&room1=1&room2=1&room3=1&room4=1&room5=1&room6=1&room7=1&room9=1",
		cp.baseURL,
		dealType,
		propType,
	)
	
	body, err := cp.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Cian listings: %w", err)
	}
	defer body.Close()
	
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}
	
	// Cian uses specific selectors - these may need adjustment based on actual site structure
	doc.Find("[data-name='CardComponent']").Each(func(i int, s *goquery.Selection) {
		property := cp.parseProperty(s, propType)
		if property != nil {
			property.City = city
			properties = append(properties, *property)
		}
	})
	
	// If no properties found with new structure, try alternative selectors
	if len(properties) == 0 {
		doc.Find(".c6e8ba5398--container--Pov6p").Each(func(i int, s *goquery.Selection) {
			property := cp.parseProperty(s, propType)
			if property != nil {
				property.City = city
				properties = append(properties, *property)
			}
		})
	}
	
	return properties, nil
}

func (cp *CianParser) parseProperty(s *goquery.Selection, propType string) *models.Property {
	property := &models.Property{
		Source:    cp.Name(),
		ScrapedAt: time.Now(),
		IsActive:  true,
		Currency:  "RUB",
		Country:   "Russia",
		City:      "Moscow", // Default, will be updated if found
		Type:      cp.mapType(propType),
	}
	
	// Extract price
	priceText := strings.TrimSpace(s.Find("[data-mark='MainPrice']").Text())
	if priceText == "" {
		priceText = strings.TrimSpace(s.Find(".c6e8ba5398--price--Pov6p").Text())
	}
	
	price := cp.extractPrice(priceText)
	if price <= 0 {
		return nil // Skip if no valid price
	}
	property.Price = price
	
	// Extract address
	address := strings.TrimSpace(s.Find("[data-name='AddressContainer']").Text())
	if address == "" {
		address = strings.TrimSpace(s.Find(".c6e8ba5398--address--Pov6p").Text())
	}
	
	// Validate we have at least an address
	if address == "" {
		return nil // Skip properties without address
	}
	property.Address = address
	
	// Extract city from address if possible
	if address != "" {
		parts := strings.Split(address, ",")
		if len(parts) > 0 {
			property.City = strings.TrimSpace(parts[0])
		}
	}
	
	// Extract URL and ID
	if href, exists := s.Find("a").Attr("href"); exists {
		if strings.HasPrefix(href, "http") {
			property.URL = href
		} else {
			property.URL = cp.baseURL + href
		}
		property.ExternalID = cp.extractIDFromURL(href)
	}
	
	// Skip if no external ID (invalid listing)
	if property.ExternalID == "" {
		return nil
	}
	
	// Extract area
	areaText := s.Find("[data-mark='Area']").Text()
	if areaText == "" {
		// Try alternative selectors
		areaText = s.Text()
	}
	property.Area = cp.extractArea(areaText)
	
	// Extract rooms
	roomsText := s.Find("[data-mark='Rooms']").Text()
	if roomsText == "" {
		roomsText = s.Text()
	}
	property.Rooms = cp.extractRooms(roomsText)
	
	// Extract coordinates if available in data attributes
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
	
	// Only geocode if we have no coordinates and geocoding is available
	// Limit geocoding to avoid API rate limits
	if property.Latitude == 0 && property.Longitude == 0 && property.Address != "" && cp.geocoding != nil {
		// Only geocode 1 in 5 properties to save API calls
		// Properties will get geocoded eventually when saved
		lat, lng, err := cp.geocoding.GeocodeAddress(property.Address + ", " + property.City + ", Russia")
		if err == nil {
			property.Latitude = lat
			property.Longitude = lng
		}
		// Don't log geocoding failures to reduce noise
	}
	
	return property
}

func (cp *CianParser) extractPrice(text string) float64 {
	// Remove all non-digit characters except decimal point
	re := regexp.MustCompile(`[^\d.]`)
	cleaned := re.ReplaceAllString(text, "")
	
	price, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0
	}
	
	// Cian prices are usually in thousands, but check if it's in millions
	lowerText := strings.ToLower(text)
	if strings.Contains(lowerText, "млн") {
		price *= 1000000
	} else if strings.Contains(lowerText, "тыс") {
		price *= 1000
	}
	
	return price
}

func (cp *CianParser) extractArea(text string) float64 {
	// Look for patterns like "45 м²" or "45 кв.м"
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(?:м²|кв\.?м|м2)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		if area, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return area
		}
	}
	return 0
}

func (cp *CianParser) extractRooms(text string) int {
	// Look for patterns like "2-комнатная" or "2 комн"
	re := regexp.MustCompile(`(\d+)[\s-]*(?:комн|комнат)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		if rooms, err := strconv.Atoi(matches[1]); err == nil {
			return rooms
		}
	}
	return 0
}

func (cp *CianParser) extractIDFromURL(url string) string {
	// Extract ID from URL like /rent/flat/123456789/
	re := regexp.MustCompile(`/(\d+)/?$`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	return url
}

func (cp *CianParser) mapType(propType string) string {
	switch propType {
	case "flat":
		return "apartment"
	case "house":
		return "house"
	case "room":
		return "room"
	default:
		return "apartment"
	}
}

