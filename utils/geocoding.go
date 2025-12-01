package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"pricemap-go/config"
)

type GeocodingService struct {
	client *http.Client
}

func NewGeocodingService() *GeocodingService {
	return &GeocodingService{
		client: &http.Client{},
	}
}

// GeocodeAddress converts an address to coordinates using OpenCage API or Nominatim
func (gs *GeocodingService) GeocodeAddress(address string) (lat, lng float64, err error) {
	if config.AppConfig.OpenCageAPIKey == "" {
		// Fallback to Nominatim (OpenStreetMap) - free but rate-limited
		return gs.geocodeWithNominatim(address)
	}
	
	return gs.geocodeWithOpenCage(address)
}

// geocodeWithOpenCage uses OpenCage Geocoding API
func (gs *GeocodingService) geocodeWithOpenCage(address string) (lat, lng float64, err error) {
	apiURL := fmt.Sprintf(
		"https://api.opencagedata.com/geocode/v1/json?q=%s&key=%s&limit=1",
		url.QueryEscape(address),
		config.AppConfig.OpenCageAPIKey,
	)
	
	resp, err := gs.client.Get(apiURL)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to geocode: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("geocoding API returned status %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}
	
	var result struct {
		Results []struct {
			Geometry struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"geometry"`
		} `json:"results"`
	}
	
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, 0, err
	}
	
	if len(result.Results) == 0 {
		return 0, 0, fmt.Errorf("no results found for address: %s", address)
	}
	
	return result.Results[0].Geometry.Lat, result.Results[0].Geometry.Lng, nil
}

// geocodeWithNominatim uses OpenStreetMap Nominatim (free, rate-limited)
func (gs *GeocodingService) geocodeWithNominatim(address string) (lat, lng float64, err error) {
	apiURL := fmt.Sprintf(
		"https://nominatim.openstreetmap.org/search?q=%s&format=json&limit=1",
		url.QueryEscape(address),
	)
	
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return 0, 0, err
	}
	
	// Nominatim requires User-Agent
	req.Header.Set("User-Agent", config.AppConfig.UserAgent)
	
	resp, err := gs.client.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to geocode: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("geocoding API returned status %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}
	
	var results []struct {
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}
	
	if err := json.Unmarshal(body, &results); err != nil {
		return 0, 0, err
	}
	
	if len(results) == 0 {
		return 0, 0, fmt.Errorf("no results found for address: %s", address)
	}
	
	// Parse coordinates
	if _, err := fmt.Sscanf(results[0].Lat, "%f", &lat); err != nil {
		return 0, 0, err
	}
	if _, err := fmt.Sscanf(results[0].Lon, "%f", &lng); err != nil {
		return 0, 0, err
	}
	
	return lat, lng, nil
}

