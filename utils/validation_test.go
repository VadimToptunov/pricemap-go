package utils

import (
	"testing"
	"pricemap-go/models"
)

func TestValidateProperty(t *testing.T) {
	tests := []struct {
		name    string
		property *models.Property
		wantErr bool
		errType error
	}{
		{
			name: "valid property",
			property: &models.Property{
				Price:      100000,
				Latitude:   55.7558,
				Longitude:  37.6173,
				Source:     "test",
				ExternalID: "123",
			},
			wantErr: false,
		},
		{
			name: "invalid price",
			property: &models.Property{
				Price:      0,
				Latitude:   55.7558,
				Longitude:  37.6173,
				Source:     "test",
				ExternalID: "123",
			},
			wantErr: true,
			errType: ErrInvalidPrice,
		},
		{
			name: "missing location",
			property: &models.Property{
				Price:      100000,
				Latitude:   0,
				Longitude:  0,
				Address:    "",
				Source:     "test",
				ExternalID: "123",
			},
			wantErr: true,
			errType: ErrMissingLocation,
		},
		{
			name: "missing source",
			property: &models.Property{
				Price:      100000,
				Latitude:   55.7558,
				Longitude:  37.6173,
				Source:     "",
				ExternalID: "123",
			},
			wantErr: true,
			errType: ErrMissingSource,
		},
		{
			name: "missing external ID",
			property: &models.Property{
				Price:      100000,
				Latitude:   55.7558,
				Longitude:  37.6173,
				Source:     "test",
				ExternalID: "",
			},
			wantErr: true,
			errType: ErrMissingExternalID,
		},
		{
			name: "valid with address",
			property: &models.Property{
				Price:      100000,
				Latitude:   0,
				Longitude:  0,
				Address:    "123 Main St",
				Source:     "test",
				ExternalID: "123",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProperty(tt.property)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProperty() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != tt.errType {
				t.Errorf("ValidateProperty() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestNormalizeProperty(t *testing.T) {
	tests := []struct {
		name     string
		property *models.Property
		expected *models.Property
	}{
		{
			name: "normalize type",
			property: &models.Property{
				Type:     "flat",
				Currency: "",
			},
			expected: &models.Property{
				Type:     "apartment",
				Currency: "USD",
			},
		},
		{
			name: "normalize house type",
			property: &models.Property{
				Type:     "home",
				Currency: "",
			},
			expected: &models.Property{
				Type:     "house",
				Currency: "USD",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			NormalizeProperty(tt.property)
			if tt.property.Type != tt.expected.Type {
				t.Errorf("NormalizeProperty() Type = %v, want %v", tt.property.Type, tt.expected.Type)
			}
			if tt.property.Currency != tt.expected.Currency {
				t.Errorf("NormalizeProperty() Currency = %v, want %v", tt.property.Currency, tt.expected.Currency)
			}
		})
	}
}

