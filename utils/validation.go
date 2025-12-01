package utils

import "pricemap-go/models"

// ValidateProperty validates a property before saving
func ValidateProperty(property *models.Property) error {
	if property.Price <= 0 {
		return ErrInvalidPrice
	}
	
	if property.Latitude == 0 && property.Longitude == 0 {
		if property.Address == "" {
			return ErrMissingLocation
		}
	}
	
	if property.Source == "" {
		return ErrMissingSource
	}
	
	if property.ExternalID == "" {
		return ErrMissingExternalID
	}
	
	return nil
}

// NormalizeProperty normalizes property data
func NormalizeProperty(property *models.Property) {
	// Normalize city name
	if property.City != "" {
		property.City = normalizeString(property.City)
	}
	
	// Normalize country name
	if property.Country != "" {
		property.Country = normalizeString(property.Country)
	}
	
	// Normalize type
	if property.Type != "" {
		property.Type = normalizeType(property.Type)
	}
	
	// Ensure currency is set
	if property.Currency == "" {
		property.Currency = "USD"
	}
}

func normalizeString(s string) string {
	// Remove extra spaces, trim
	// Can add more normalization logic
	return s
}

func normalizeType(t string) string {
	// Normalize property types
	switch t {
	case "flat", "apartment", "apt":
		return "apartment"
	case "house", "home", "villa":
		return "house"
	case "room", "bedroom":
		return "room"
	default:
		return t
	}
}

// Errors
var (
	ErrInvalidPrice      = &ValidationError{Field: "price", Message: "price must be greater than 0"}
	ErrMissingLocation   = &ValidationError{Field: "location", Message: "latitude/longitude or address required"}
	ErrMissingSource     = &ValidationError{Field: "source", Message: "source is required"}
	ErrMissingExternalID = &ValidationError{Field: "external_id", Message: "external_id is required"}
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

