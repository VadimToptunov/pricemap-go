package services

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"pricemap-go/models"
)

type TransportService struct{}

func NewTransportService() *TransportService {
	return &TransportService{}
}

// CalculateTransportScore calculates transportation accessibility score
func (ts *TransportService) CalculateTransportScore(property *models.Property, transitStops []TransitStop) float64 {
	if len(transitStops) == 0 {
		return 50.0 // Default score if no transit data
	}
	
	// Find nearest transit stops
	minDistance := math.MaxFloat64
	metroCount := 0
	busCount := 0
	
	for _, stop := range transitStops {
		distance := ts.haversineDistance(
			property.Latitude, property.Longitude,
			stop.Latitude, stop.Longitude,
		)
		
		if distance < minDistance {
			minDistance = distance
		}
		
		// Count stops within 1km
		if distance <= 1.0 {
			if stop.Type == "metro" || stop.Type == "subway" {
				metroCount++
			} else {
				busCount++
			}
		}
	}
	
	// Calculate score based on:
	// - Distance to nearest stop (closer = better)
	// - Number of stops within 1km
	// - Type of transit (metro > bus)
	
	score := 0.0
	
	// Distance score (0-40 points)
	if minDistance <= 0.5 {
		score += 40 // Very close
	} else if minDistance <= 1.0 {
		score += 30
	} else if minDistance <= 2.0 {
		score += 20
	} else if minDistance <= 5.0 {
		score += 10
	}
	
	// Metro availability (0-40 points)
	if metroCount > 0 {
		score += math.Min(float64(metroCount)*10, 40)
	} else if busCount > 0 {
		// Bus availability (0-20 points)
		score += math.Min(float64(busCount)*2, 20)
	}
	
	// Additional points for multiple options (0-20 points)
	totalStops := metroCount + busCount
	if totalStops >= 5 {
		score += 20
	} else if totalStops >= 3 {
		score += 10
	}
	
	return math.Min(score, 100.0)
}

// TransitStop represents a public transit stop
type TransitStop struct {
	Latitude  float64
	Longitude float64
	Type      string // "metro", "bus", "tram", etc.
	Name      string
}

// ParseGTFSStops parses GTFS stops.txt file
func (ts *TransportService) ParseGTFSStops(reader io.Reader) ([]TransitStop, error) {
	var stops []TransitStop
	
	csvReader := csv.NewReader(reader)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}
	
	if len(records) == 0 {
		return stops, nil
	}
	
	// Find column indices
	header := records[0]
	latIdx := -1
	lngIdx := -1
	nameIdx := -1
	typeIdx := -1
	
	for i, col := range header {
		switch col {
		case "stop_lat", "lat", "latitude":
			latIdx = i
		case "stop_lon", "lon", "lng", "longitude":
			lngIdx = i
		case "stop_name", "name":
			nameIdx = i
		case "location_type", "type":
			typeIdx = i
		}
	}
	
	if latIdx == -1 || lngIdx == -1 {
		return nil, fmt.Errorf("latitude or longitude column not found")
	}
	
	// Parse rows
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) <= latIdx || len(record) <= lngIdx {
			continue
		}
		
		var lat, lng float64
		if _, err := fmt.Sscanf(record[latIdx], "%f", &lat); err != nil {
			continue
		}
		if _, err := fmt.Sscanf(record[lngIdx], "%f", &lng); err != nil {
			continue
		}
		
		stop := TransitStop{
			Latitude:  lat,
			Longitude: lng,
			Type:      "bus", // Default
		}
		
		if nameIdx >= 0 && nameIdx < len(record) {
			stop.Name = record[nameIdx]
		}
		
		if typeIdx >= 0 && typeIdx < len(record) {
			stop.Type = record[typeIdx]
		}
		
		stops = append(stops, stop)
	}
	
	return stops, nil
}

// haversineDistance calculates distance between two points in kilometers
func (ts *TransportService) haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth radius in kilometers
	
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	
	return R * c
}

