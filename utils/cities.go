package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"pricemap-go/config"
)

// City represents a city/location
type City struct {
	Name      string
	Country   string
	Latitude  float64
	Longitude float64
	Population int
}

// CityService provides city data
type CityService struct {
	client *http.Client
}

func NewCityService() *CityService {
	return &CityService{
		client: &http.Client{},
	}
}

// GetCitiesFromOpenStreetMap gets cities from OpenStreetMap Nominatim
// This can be used to find all cities in a country or region
func (cs *CityService) GetCitiesFromOpenStreetMap(countryCode string, minPopulation int) ([]City, error) {
	// Query Nominatim for cities in a country
	query := fmt.Sprintf(
		"https://nominatim.openstreetmap.org/search?country=%s&featuretype=city&format=json&limit=1000&addressdetails=1",
		countryCode,
	)
	
	req, err := http.NewRequest("GET", query, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("User-Agent", config.AppConfig.UserAgent)
	
	resp, err := cs.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var results []struct {
		DisplayName string `json:"display_name"`
		Lat         string  `json:"lat"`
		Lon         string  `json:"lon"`
		Address     struct {
			City      string `json:"city"`
			Town      string `json:"town"`
			Village   string `json:"village"`
			Country   string `json:"country"`
		} `json:"address"`
		Extratags struct {
			Population string `json:"population"`
		} `json:"extratags"`
	}
	
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, err
	}
	
	var cities []City
	seen := make(map[string]bool)
	
	for _, result := range results {
		var cityName string
		if result.Address.City != "" {
			cityName = result.Address.City
		} else if result.Address.Town != "" {
			cityName = result.Address.Town
		} else if result.Address.Village != "" {
			cityName = result.Address.Village
		} else {
			continue
		}
		
		// Create unique key
		key := fmt.Sprintf("%s_%s", cityName, result.Address.Country)
		if seen[key] {
			continue
		}
		seen[key] = true
		
		var lat, lon float64
		fmt.Sscanf(result.Lat, "%f", &lat)
		fmt.Sscanf(result.Lon, "%f", &lon)
		
		population := 0
		if result.Extratags.Population != "" {
			fmt.Sscanf(result.Extratags.Population, "%d", &population)
		}
		
		if population < minPopulation {
			continue
		}
		
		cities = append(cities, City{
			Name:       cityName,
			Country:    result.Address.Country,
			Latitude:   lat,
			Longitude:  lon,
			Population: population,
		})
	}
	
	return cities, nil
}

// GetMajorCities returns a list of major cities worldwide
// This is a curated list that can be expanded
func GetMajorCities() []City {
	return []City{
		// North America
		{"New York", "United States", 40.7128, -74.0060, 8336817},
		{"Los Angeles", "United States", 34.0522, -118.2437, 3971883},
		{"Chicago", "United States", 41.8781, -87.6298, 2693976},
		{"Houston", "United States", 29.7604, -95.3698, 2320268},
		{"Phoenix", "United States", 33.4484, -112.0740, 1680992},
		{"Toronto", "Canada", 43.6532, -79.3832, 2930000},
		{"Vancouver", "Canada", 49.2827, -123.1207, 675218},
		{"Montreal", "Canada", 45.5017, -73.5673, 1780000},
		
		// Europe
		{"London", "United Kingdom", 51.5074, -0.1278, 9002488},
		{"Paris", "France", 48.8566, 2.3522, 2161000},
		{"Berlin", "Germany", 52.5200, 13.4050, 3669491},
		{"Madrid", "Spain", 40.4168, -3.7038, 3223334},
		{"Rome", "Italy", 41.9028, 12.4964, 2873000},
		{"Amsterdam", "Netherlands", 52.3676, 4.9041, 872680},
		{"Barcelona", "Spain", 41.3851, 2.1734, 1636762},
		{"Moscow", "Russia", 55.7558, 37.6173, 12615279},
		{"Saint Petersburg", "Russia", 59.9343, 30.3351, 5398064},
		{"Istanbul", "Turkey", 41.0082, 28.9784, 15519267},
		
		// Asia
		{"Tokyo", "Japan", 35.6762, 139.6503, 13929286},
		{"Shanghai", "China", 31.2304, 121.4737, 24870895},
		{"Beijing", "China", 39.9042, 116.4074, 21540000},
		{"Mumbai", "India", 19.0760, 72.8777, 12478447},
		{"Delhi", "India", 28.6139, 77.2090, 32900000},
		{"Bangkok", "Thailand", 13.7563, 100.5018, 10539000},
		{"Singapore", "Singapore", 1.3521, 103.8198, 5685807},
		{"Seoul", "South Korea", 37.5665, 126.9780, 9720846},
		{"Dubai", "United Arab Emirates", 25.2048, 55.2708, 3400000},
		
		// Oceania
		{"Sydney", "Australia", -33.8688, 151.2093, 5312163},
		{"Melbourne", "Australia", -37.8136, 144.9631, 5078193},
		{"Auckland", "New Zealand", -36.8485, 174.7633, 1657000},
		
		// South America
		{"São Paulo", "Brazil", -23.5505, -46.6333, 12325232},
		{"Rio de Janeiro", "Brazil", -22.9068, -43.1729, 6747815},
		{"Buenos Aires", "Argentina", -34.6037, -58.3816, 3075646},
		{"Lima", "Peru", -12.0464, -77.0428, 9750000},
		
		// Africa
		{"Cairo", "Egypt", 30.0444, 31.2357, 10230350},
		{"Lagos", "Nigeria", 6.5244, 3.3792, 15388000},
		{"Johannesburg", "South Africa", -26.2041, 28.0473, 5634800},
		{"Cape Town", "South Africa", -33.9249, 18.4241, 4618000},
	}
}

// GetCitiesByCountry returns major cities for a specific country
func GetCitiesByCountry(country string) []City {
	allCities := GetMajorCities()
	var result []City
	for _, city := range allCities {
		if city.Country == country {
			result = append(result, city)
		}
	}
	return result
}

