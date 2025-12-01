package utils

import (
	"testing"
	"pricemap-go/models"
)

func TestCurrencyConverter_Convert(t *testing.T) {
	cc := NewCurrencyConverter()

	tests := []struct {
		name     string
		price    float64
		from     string
		to       string
		wantErr  bool
		expected float64
	}{
		{
			name:     "USD to USD",
			price:    1000,
			from:     "USD",
			to:       "USD",
			wantErr:  false,
			expected: 1000,
		},
		{
			name:     "USD to EUR",
			price:    1000,
			from:     "USD",
			to:       "EUR",
			wantErr:  false,
			expected: 920, // 1000 * 0.92
		},
		{
			name:     "EUR to USD",
			price:    920,
			from:     "EUR",
			to:       "USD",
			wantErr:  false,
			expected: 1000, // 920 / 0.92
		},
		{
			name:    "unknown currency",
			price:   1000,
			from:    "XXX",
			to:      "USD",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cc.Convert(tt.price, tt.from, tt.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Allow small floating point differences
				diff := got - tt.expected
				if diff < 0 {
					diff = -diff
				}
				if diff > 0.01 {
					t.Errorf("Convert() = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}

func TestCurrencyConverter_NormalizeToUSD(t *testing.T) {
	cc := NewCurrencyConverter()

	tests := []struct {
		name     string
		property *models.Property
		wantErr  bool
	}{
		{
			name: "already USD",
			property: &models.Property{
				Price:    1000,
				Currency: "USD",
			},
			wantErr: false,
		},
		{
			name: "convert EUR to USD",
			property: &models.Property{
				Price:    920,
				Currency: "EUR",
			},
			wantErr: false,
		},
		{
			name: "unknown currency",
			property: &models.Property{
				Price:    1000,
				Currency: "XXX",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalPrice := tt.property.Price
			originalCurrency := tt.property.Currency

			err := cc.NormalizeToUSD(tt.property)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeToUSD() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.property.Currency != "USD" {
					t.Errorf("NormalizeToUSD() Currency = %v, want USD", tt.property.Currency)
				}
				if originalCurrency == "USD" && tt.property.Price != originalPrice {
					t.Errorf("NormalizeToUSD() should not change USD price")
				}
			} else {
				// Should not modify on error
				if tt.property.Price != originalPrice || tt.property.Currency != originalCurrency {
					t.Errorf("NormalizeToUSD() should not modify property on error")
				}
			}
		})
	}
}

