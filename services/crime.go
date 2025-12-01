package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"pricemap-go/config"
	"pricemap-go/models"
)

type CrimeService struct {
	client *http.Client
}

func NewCrimeService() *CrimeService {
	return &CrimeService{
		client: &http.Client{},
	}
}

// GetCrimeData fetches crime data for a location
func (cs *CrimeService) GetCrimeData(lat, lng float64, country, city string) (*CrimeData, error) {
	// Try different sources based on country
	switch country {
	case "United States":
		return cs.getUSCrimeData(city)
	case "United Kingdom":
		return cs.getUKCrimeData(lat, lng)
	default:
		// Generic fallback - use OpenStreetMap or other sources
		return cs.getGenericCrimeData(lat, lng, city)
	}
}

// CrimeData represents crime statistics for an area
type CrimeData struct {
	Score      float64 `json:"score"`      // 0-100, where 100 is safest
	CrimeRate  float64 `json:"crime_rate"` // Crimes per 1000 people
	ViolentCrime float64 `json:"violent_crime"`
	PropertyCrime float64 `json:"property_crime"`
	Source     string  `json:"source"`
	LastUpdated string `json:"last_updated"`
}

// getUSCrimeData fetches crime data from US sources
func (cs *CrimeService) getUSCrimeData(city string) (*CrimeData, error) {
	// Try city open data portals
	// Example: Chicago, NYC, LA have open data APIs
	
	// For now, return placeholder - can be extended with actual API calls
	return &CrimeData{
		Score:      70.0, // Default safe score
		CrimeRate:  30.0, // Default crime rate
		Source:     "placeholder",
		LastUpdated: "2024-01-01",
	}, nil
}

// getUKCrimeData fetches crime data from UK Police API
func (cs *CrimeService) getUKCrimeData(lat, lng float64) (*CrimeData, error) {
	// UK Police Data API
	url := fmt.Sprintf("https://data.police.uk/api/crimes-street/all-crime?lat=%.6f&lng=%.6f",
		lat, lng)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("User-Agent", config.AppConfig.UserAgent)
	
	resp, err := cs.client.Do(req)
	if err != nil {
		// Fallback to default
		return &CrimeData{Score: 70.0, Source: "fallback"}, nil
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return &CrimeData{Score: 70.0, Source: "fallback"}, nil
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &CrimeData{Score: 70.0, Source: "fallback"}, nil
	}
	
	var crimes []map[string]interface{}
	if err := json.Unmarshal(body, &crimes); err != nil {
		return &CrimeData{Score: 70.0, Source: "fallback"}, nil
	}
	
	// Calculate crime score based on number of crimes
	crimeCount := len(crimes)
	score := 100.0
	
	// Adjust score based on crime count
	// More crimes = lower score
	if crimeCount > 100 {
		score = 30.0
	} else if crimeCount > 50 {
		score = 50.0
	} else if crimeCount > 20 {
		score = 70.0
	} else if crimeCount > 10 {
		score = 85.0
	}
	
	return &CrimeData{
		Score:      score,
		CrimeRate:  float64(crimeCount) * 10, // Approximate
		Source:     "uk-police-api",
		LastUpdated: "2024-01-01",
	}, nil
}

// getGenericCrimeData gets generic crime data
func (cs *CrimeService) getGenericCrimeData(lat, lng float64, city string) (*CrimeData, error) {
	// Placeholder - can integrate with other sources
	// For example: Numbeo crime index, local police APIs, etc.
	
	return &CrimeData{
		Score:      70.0,
		CrimeRate:  30.0,
		Source:     "generic",
		LastUpdated: "2024-01-01",
	}, nil
}

// CalculateCrimeScore calculates crime score for a property
func (cs *CrimeService) CalculateCrimeScore(property *models.Property) (float64, string, error) {
	crimeData, err := cs.GetCrimeData(
		property.Latitude,
		property.Longitude,
		property.Country,
		property.City,
	)
	
	if err != nil {
		return 70.0, "{}", err
	}
	
	dataJSON, _ := json.Marshal(crimeData)
	
	return crimeData.Score, string(dataJSON), nil
}

