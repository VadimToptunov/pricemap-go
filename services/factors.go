package services

import (
	"encoding/json"
	"log"
	"math"
	"pricemap-go/database"
	"pricemap-go/models"
)

type FactorsService struct{}

func NewFactorsService() *FactorsService {
	return &FactorsService{}
}

// CalculateFactors calculates all factors for a property
func (fs *FactorsService) CalculateFactors(property *models.Property) (*models.PropertyFactors, error) {
	factors := &models.PropertyFactors{
		PropertyID: property.ID,
	}
	
	// Calculate crime score
	crimeScore, crimeData, err := fs.calculateCrimeScore(property)
	if err != nil {
		log.Printf("Error calculating crime score: %v", err)
	} else {
		factors.CrimeScore = crimeScore
		factors.CrimeData = crimeData
	}
	
	// Calculate transportation accessibility
	transportScore, transportData, err := fs.calculateTransportScore(property)
	if err != nil {
		log.Printf("Error calculating transport score: %v", err)
	} else {
		factors.TransportScore = transportScore
		factors.TransportData = transportData
	}
	
	// Calculate education score
	educationScore, educationData, err := fs.calculateEducationScore(property)
	if err != nil {
		log.Printf("Error calculating education score: %v", err)
	} else {
		factors.EducationScore = educationScore
		factors.EducationData = educationData
	}
	
	// Calculate infrastructure score
	infraScore, infraData, err := fs.calculateInfrastructureScore(property)
	if err != nil {
		log.Printf("Error calculating infrastructure score: %v", err)
	} else {
		factors.InfrastructureScore = infraScore
		factors.InfrastructureData = infraData
	}
	
	// Calculate overall rating (weighted sum)
	factors.OverallScore = fs.calculateOverallScore(factors)
	
	return factors, nil
}

// calculateCrimeScore calculates safety rating (0-100)
func (fs *FactorsService) calculateCrimeScore(property *models.Property) (float64, string, error) {
	crimeService := NewCrimeService()
	return crimeService.CalculateCrimeScore(property)
}

// calculateTransportScore calculates transportation accessibility (0-100)
func (fs *FactorsService) calculateTransportScore(property *models.Property) (float64, string, error) {
	// TODO: Integration with Google Maps API, OpenStreetMap, GTFS
	// For now, use a basic calculation based on location
	
	transportService := NewTransportService()
	
	// In a real implementation, you would:
	// 1. Load GTFS data for the city
	// 2. Find nearest transit stops
	// 3. Calculate score based on distance and availability
	
	// Placeholder: calculate based on city center distance (if coordinates available)
	score := 65.0 // Default
	
	// If we have coordinates, we could calculate distance to known transit hubs
	// For now, return default score
	
	transportData := map[string]interface{}{
		"metro_stations": []map[string]interface{}{},
		"bus_stops": []map[string]interface{}{},
		"walking_time_to_transit": 0,
		"note": "Full integration with GTFS data pending",
	}
	
	dataJSON, _ := json.Marshal(transportData)
	
	// Use transport service for calculation (when transit data is available)
	_ = transportService // Will be used when GTFS data is loaded
	
	return score, string(dataJSON), nil
}

// calculateEducationScore calculates education rating (0-100)
func (fs *FactorsService) calculateEducationScore(property *models.Property) (float64, string, error) {
	educationService := NewEducationService()
	return educationService.CalculateEducationScore(property)
}

// calculateInfrastructureScore calculates infrastructure rating (0-100)
func (fs *FactorsService) calculateInfrastructureScore(property *models.Property) (float64, string, error) {
	// TODO: Integration with POI data (Points of Interest)
	
	infraData := map[string]interface{}{
		"shops": 0,
		"parks": 0,
		"hospitals": 0,
		"restaurants": 0,
	}
	
	dataJSON, _ := json.Marshal(infraData)
	
	score := 70.0
	
	return score, string(dataJSON), nil
}

// calculateOverallScore calculates overall rating
func (fs *FactorsService) calculateOverallScore(factors *models.PropertyFactors) float64 {
	// Weighted sum of all factors
	weights := map[string]float64{
		"crime":        0.25,
		"transport":    0.25,
		"education":    0.20,
		"infrastructure": 0.30,
	}
	
	overall := factors.CrimeScore*weights["crime"] +
		factors.TransportScore*weights["transport"] +
		factors.EducationScore*weights["education"] +
		factors.InfrastructureScore*weights["infrastructure"]
	
	return math.Round(overall*100) / 100
}

// SaveFactors saves factors to database
func (fs *FactorsService) SaveFactors(factors *models.PropertyFactors) error {
	var existing models.PropertyFactors
	result := database.DB.Where("property_id = ?", factors.PropertyID).First(&existing)
	
	if result.Error == nil {
		// Update existing
		factors.ID = existing.ID
		return database.DB.Save(factors).Error
	}
	
	// Create new
	return database.DB.Create(factors).Error
}

