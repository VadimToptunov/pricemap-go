package utils

import (
	"testing"
)

func TestGetMajorCities(t *testing.T) {
	cities := GetMajorCities()

	if len(cities) == 0 {
		t.Errorf("GetMajorCities() returned empty list")
	}

	// Check that we have cities from different continents
	countries := make(map[string]bool)
	for _, city := range cities {
		countries[city.Country] = true
	}

	if len(countries) < 5 {
		t.Errorf("GetMajorCities() should include cities from multiple countries, got %d", len(countries))
	}
}

func TestGetCitiesByCountry(t *testing.T) {
	tests := []struct {
		country string
		want    int
	}{
		{"United States", 5}, // Should have at least 5 US cities
		{"Russia", 2},         // Should have at least 2 Russian cities
		{"United Kingdom", 1}, // Should have at least 1 UK city
		{"Nonexistent", 0},    // Should return empty for non-existent country
	}

	for _, tt := range tests {
		t.Run(tt.country, func(t *testing.T) {
			cities := GetCitiesByCountry(tt.country)
			if len(cities) < tt.want {
				t.Errorf("GetCitiesByCountry(%s) = %d cities, want at least %d", tt.country, len(cities), tt.want)
			}
		})
	}
}

func TestCity_Structure(t *testing.T) {
	cities := GetMajorCities()

	for _, city := range cities {
		if city.Name == "" {
			t.Errorf("City name should not be empty")
		}
		if city.Country == "" {
			t.Errorf("City country should not be empty")
		}
		if city.Latitude == 0 && city.Longitude == 0 {
			t.Errorf("City should have coordinates: %s", city.Name)
		}
		if city.Latitude < -90 || city.Latitude > 90 {
			t.Errorf("City latitude out of range: %s, lat=%f", city.Name, city.Latitude)
		}
		if city.Longitude < -180 || city.Longitude > 180 {
			t.Errorf("City longitude out of range: %s, lng=%f", city.Name, city.Longitude)
		}
	}
}

