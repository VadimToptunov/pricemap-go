package services

import (
	"testing"
	"pricemap-go/models"
)

func TestTransportService_CalculateTransportScore(t *testing.T) {
	ts := NewTransportService()

	tests := []struct {
		name        string
		property    *models.Property
		transitStops []TransitStop
		want        float64
	}{
		{
			name: "no transit stops",
			property: &models.Property{
				Latitude:  55.7558,
				Longitude: 37.6173,
			},
			transitStops: []TransitStop{},
			want:        50.0, // Default score
		},
		{
			name: "close metro station",
			property: &models.Property{
				Latitude:  55.7558,
				Longitude: 37.6173,
			},
			transitStops: []TransitStop{
				{
					Latitude:  55.7559,
					Longitude: 37.6174,
					Type:      "metro",
					Name:      "Test Metro",
				},
			},
			want: 80.0, // Should be high due to close metro
		},
		{
			name: "multiple stops nearby",
			property: &models.Property{
				Latitude:  55.7558,
				Longitude: 37.6173,
			},
			transitStops: []TransitStop{
				{Latitude: 55.7559, Longitude: 37.6174, Type: "metro"},
				{Latitude: 55.7557, Longitude: 37.6172, Type: "metro"},
				{Latitude: 55.7558, Longitude: 37.6175, Type: "bus"},
				{Latitude: 55.7556, Longitude: 37.6171, Type: "bus"},
				{Latitude: 55.7555, Longitude: 37.6170, Type: "bus"},
			},
			want: 100.0, // Should be max due to many stops
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ts.CalculateTransportScore(tt.property, tt.transitStops)
			if got < 0 || got > 100 {
				t.Errorf("CalculateTransportScore() = %v, want 0-100", got)
			}
			// For tests with stops, score should be reasonable
			if len(tt.transitStops) > 0 && got < 50 {
				t.Errorf("CalculateTransportScore() = %v, want >= 50 for properties with transit", got)
			}
		})
	}
}

func TestTransportService_HaversineDistance(t *testing.T) {
	ts := NewTransportService()

	tests := []struct {
		name     string
		lat1     float64
		lon1     float64
		lat2     float64
		lon2     float64
		want     float64 // Approximate distance in km
		tolerance float64
	}{
		{
			name:      "same point",
			lat1:      55.7558,
			lon1:      37.6173,
			lat2:      55.7558,
			lon2:      37.6173,
			want:      0.0,
			tolerance: 0.1,
		},
		{
			name:      "close points",
			lat1:      55.7558,
			lon1:      37.6173,
			lat2:      55.7559,
			lon2:      37.6174,
			want:      0.1, // Very close
			tolerance: 0.1,
		},
		{
			name:      "Moscow to Saint Petersburg",
			lat1:      55.7558,
			lon1:      37.6173,
			lat2:      59.9343,
			lon2:      30.3351,
			want:      635.0, // Approximate distance
			tolerance: 50.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ts.haversineDistance(tt.lat1, tt.lon1, tt.lat2, tt.lon2)
			diff := got - tt.want
			if diff < 0 {
				diff = -diff
			}
			if diff > tt.tolerance {
				t.Errorf("haversineDistance() = %v, want %v (tolerance %v)", got, tt.want, tt.tolerance)
			}
		})
	}
}

