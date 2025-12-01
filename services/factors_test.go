package services

import (
	"testing"
	"pricemap-go/models"
	"time"
)

func TestFactorsService_CalculateOverallScore(t *testing.T) {
	fs := NewFactorsService()

	tests := []struct {
		name    string
		factors *models.PropertyFactors
		want    float64
	}{
		{
			name: "all factors equal",
			factors: &models.PropertyFactors{
				CrimeScore:         80.0,
				TransportScore:     80.0,
				EducationScore:     80.0,
				InfrastructureScore: 80.0,
			},
			want: 80.0,
		},
		{
			name: "mixed factors",
			factors: &models.PropertyFactors{
				CrimeScore:         100.0,
				TransportScore:     50.0,
				EducationScore:     50.0,
				InfrastructureScore: 50.0,
			},
			want: 62.5, // 100*0.25 + 50*0.25 + 50*0.20 + 50*0.30
		},
		{
			name: "low scores",
			factors: &models.PropertyFactors{
				CrimeScore:         20.0,
				TransportScore:     20.0,
				EducationScore:     20.0,
				InfrastructureScore: 20.0,
			},
			want: 20.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fs.calculateOverallScore(tt.factors)
			diff := got - tt.want
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.01 {
				t.Errorf("calculateOverallScore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFactorsService_CalculateFactors(t *testing.T) {
	fs := NewFactorsService()

	property := &models.Property{
		ID:        1,
		Latitude:  55.7558,
		Longitude: 37.6173,
		Country:   "Russia",
		City:      "Moscow",
		ScrapedAt: time.Now(),
	}

	factors, err := fs.CalculateFactors(property)
	if err != nil {
		t.Fatalf("CalculateFactors() error = %v", err)
	}

	if factors.PropertyID != property.ID {
		t.Errorf("CalculateFactors() PropertyID = %v, want %v", factors.PropertyID, property.ID)
	}

	if factors.CrimeScore < 0 || factors.CrimeScore > 100 {
		t.Errorf("CalculateFactors() CrimeScore = %v, want 0-100", factors.CrimeScore)
	}

	if factors.TransportScore < 0 || factors.TransportScore > 100 {
		t.Errorf("CalculateFactors() TransportScore = %v, want 0-100", factors.TransportScore)
	}

	if factors.OverallScore < 0 || factors.OverallScore > 100 {
		t.Errorf("CalculateFactors() OverallScore = %v, want 0-100", factors.OverallScore)
	}
}

