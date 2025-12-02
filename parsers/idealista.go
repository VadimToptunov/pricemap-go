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

// IdealistaParser parses real estate data from idealista.com (Spain)
type IdealistaParser struct {
	*BaseParser
	geocoding *utils.GeocodingService
}

func NewIdealistaParser() *IdealistaParser {
	return &IdealistaParser{
		BaseParser: NewBaseParser("https://www.idealista.com"),
		geocoding:  utils.NewGeocodingService(),
	}
}

func (ip *IdealistaParser) Name() string {
	return "idealista"
}

func (ip *IdealistaParser) Parse(ctx context.Context) ([]models.Property, error) {
	var allProperties []models.Property
	
	// Spanish cities
	cities := []string{
		"Madrid", "Barcelona", "Valencia", "Seville", "Zaragoza",
		"Málaga", "Murcia", "Palma", "Las Palmas", "Bilbao",
		"Alicante", "Córdoba", "Valladolid", "Vigo", "Gijón",
		"Granada", "Vitoria", "A Coruña", "Elche", "Santa Cruz de Tenerife",
	}
	
	// Parse for both sale and rent
	dealTypes := []struct {
		path string
		name string
	}{
		{"venta-viviendas", "sale"},
		{"alquiler-viviendas", "rent"},
	}
	
	for _, city := range cities {
		for _, dealType := range dealTypes {
			properties, err := ip.parseCity(ctx, city, dealType.path, dealType.name)
			if err != nil {
				log.Printf("Error parsing %s/%s from Idealista: %v", city, dealType.name, err)
				continue
			}
			allProperties = append(allProperties, properties...)
			
			time.Sleep(2 * time.Second)
		}
	}
	
	log.Printf("Parsed %d properties from Idealista", len(allProperties))
	return allProperties, nil
}

func (ip *IdealistaParser) parseCity(ctx context.Context, city, path string, _ string) ([]models.Property, error) {
	var properties []models.Property
	
	url := fmt.Sprintf("%s/%s/%s/", ip.baseURL, path, strings.ToLower(city))
	
	body, err := ip.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Idealista listings: %w", err)
	}
	defer body.Close()
	
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}
	
	doc.Find(".item").Each(func(i int, s *goquery.Selection) {
		property := ip.parseProperty(s, city)
		if property != nil {
			properties = append(properties, *property)
		}
	})
	
	return properties, nil
}

func (ip *IdealistaParser) parseProperty(s *goquery.Selection, city string) *models.Property {
	property := &models.Property{
		Source:    ip.Name(),
		ScrapedAt: time.Now(),
		IsActive:  true,
		Currency:  "EUR",
		Country:   "Spain",
		City:      city,
		Type:      "apartment",
	}
	
	// Extract price
	priceText := strings.TrimSpace(s.Find(".item-price").Text())
	price := ip.extractPrice(priceText)
	if price <= 0 {
		return nil
	}
	property.Price = price
	
	// Extract address
	address := strings.TrimSpace(s.Find(".item-detail").Text())
	property.Address = address
	
	// Extract URL
	if href, exists := s.Find("a").Attr("href"); exists {
		if strings.HasPrefix(href, "http") {
			property.URL = href
		} else {
			property.URL = ip.baseURL + href
		}
		property.ExternalID = ip.extractIDFromURL(href)
	}
	
	// Extract details
	detailsText := s.Find(".item-detail-char").Text()
	property.Area = ip.extractArea(detailsText)
	property.Rooms = ip.extractRooms(detailsText)
	
	// Geocode address
	if property.Address != "" {
		lat, lng, err := ip.geocoding.GeocodeAddress(property.Address + ", " + city + ", Spain")
		if err == nil {
			property.Latitude = lat
			property.Longitude = lng
		}
	}
	
	return property
}

func (ip *IdealistaParser) extractPrice(text string) float64 {
	re := regexp.MustCompile(`[^\d.]`)
	cleaned := re.ReplaceAllString(text, "")
	
	price, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0
	}
	
	return price
}

func (ip *IdealistaParser) extractArea(text string) float64 {
	re := regexp.MustCompile(`(\d+)\s*(?:m²|m2)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		if area, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return area
		}
	}
	return 0
}

func (ip *IdealistaParser) extractRooms(text string) int {
	re := regexp.MustCompile(`(\d+)\s*(?:hab|dorm)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		if rooms, err := strconv.Atoi(matches[1]); err == nil {
			return rooms
		}
	}
	return 0
}

func (ip *IdealistaParser) extractIDFromURL(url string) string {
	re := regexp.MustCompile(`/(\d+)/?$`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	return url
}

