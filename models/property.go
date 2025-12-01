package models

import (
	"time"
	"gorm.io/gorm"
)

// Property represents a real estate property
type Property struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Basic information
	Source      string    `gorm:"not null;index" json:"source"` // Data source
	ExternalID  string    `gorm:"uniqueIndex:idx_source_external" json:"external_id"`
	URL         string    `gorm:"type:text" json:"url"`
	
	// Location
	Country     string    `gorm:"not null;index" json:"country"`
	City        string    `gorm:"not null;index" json:"city"`
	District    string    `gorm:"index" json:"district"`
	Address     string    `json:"address"`
	Latitude    float64   `gorm:"not null;index" json:"latitude"`
	Longitude   float64   `gorm:"not null;index" json:"longitude"`
	
	// Characteristics
	Type        string    `gorm:"not null" json:"type"` // apartment, house, etc.
	Price        float64   `gorm:"not null;index" json:"price"`
	Currency     string    `gorm:"default:'USD'" json:"currency"`
	Area         float64   `json:"area"` // area in m²
	Rooms        int       `json:"rooms"`
	Bedrooms     int       `json:"bedrooms"`
	Bathrooms    int       `json:"bathrooms"`
	Floor        int       `json:"floor"`
	TotalFloors  int       `json:"total_floors"`
	YearBuilt    int       `json:"year_built"`
	
	// Additional information
	Description  string    `gorm:"type:text" json:"description"`
	Images       []string  `gorm:"type:text[]" json:"images"`
	
	// Metadata
	ScrapedAt    time.Time `gorm:"not null" json:"scraped_at"`
	IsActive     bool      `gorm:"default:true;index" json:"is_active"`
	
	// Relations
	Factors      PropertyFactors `gorm:"foreignKey:PropertyID" json:"factors"`
}

// PropertyFactors contains factors affecting the price
type PropertyFactors struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	
	PropertyID uint      `gorm:"uniqueIndex;not null" json:"property_id"`
	
	// Crime (0-100, where 100 is the safest)
	CrimeScore      float64 `gorm:"default:0" json:"crime_score"`
	CrimeData       string  `gorm:"type:jsonb" json:"crime_data"` // Detailed data
	
	// Transportation accessibility (0-100, where 100 is excellent accessibility)
	TransportScore  float64 `gorm:"default:0" json:"transport_score"`
	TransportData   string  `gorm:"type:jsonb" json:"transport_data"` // Metro, buses, distance
	
	// Education (0-100, where 100 is the best schools)
	EducationScore  float64 `gorm:"default:0" json:"education_score"`
	EducationData   string  `gorm:"type:jsonb" json:"education_data"` // School ratings
	
	// Infrastructure
	InfrastructureScore float64 `gorm:"default:0" json:"infrastructure_score"`
	InfrastructureData  string  `gorm:"type:jsonb" json:"infrastructure_data"` // Shops, parks, hospitals
	
	// Overall rating
	OverallScore    float64 `gorm:"default:0;index" json:"overall_score"`
	
	// Additional factors
	AirQuality      float64 `json:"air_quality"`
	NoiseLevel      float64 `json:"noise_level"`
	Walkability     float64 `json:"walkability"`
}

// PriceHeatmapPoint represents a point for heatmap
type PriceHeatmapPoint struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
	Price     float64 `json:"price"`
	Score     float64 `json:"score"` // Overall rating
	Count     int     `json:"count"` // Number of properties in this area
}

