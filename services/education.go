package services

import (
	"encoding/json"
	"net/http"
	"pricemap-go/models"
)

type EducationService struct {
	client *http.Client
}

func NewEducationService() *EducationService {
	return &EducationService{
		client: &http.Client{},
	}
}

// GetEducationData fetches education/school data for a location
func (es *EducationService) GetEducationData(lat, lng float64, country, city string) (*EducationData, error) {
	switch country {
	case "United States":
		return es.getUSEducationData(lat, lng, city)
	case "United Kingdom":
		return es.getUKEducationData(lat, lng, city)
	default:
		return es.getGenericEducationData(lat, lng, city)
	}
}

// EducationData represents education statistics
type EducationData struct {
	Score         float64 `json:"score"`          // 0-100
	AverageRating float64 `json:"average_rating"` // Average school rating
	SchoolsCount  int     `json:"schools_count"`  // Number of schools nearby
	TopSchoolRating float64 `json:"top_school_rating"`
	Source        string  `json:"source"`
}

// getUSEducationData fetches from GreatSchools or NCES
func (es *EducationService) getUSEducationData(lat, lng float64, city string) (*EducationData, error) {
	// GreatSchools API or NCES data
	// For now, placeholder
	
	return &EducationData{
		Score:         65.0,
		AverageRating: 7.0,
		SchoolsCount:  5,
		Source:        "placeholder",
	}, nil
}

// getUKEducationData fetches from Ofsted
func (es *EducationService) getUKEducationData(lat, lng float64, city string) (*EducationData, error) {
	// UK Ofsted data or Department for Education
	// For now, placeholder
	
	return &EducationData{
		Score:         70.0,
		AverageRating: 3.5, // Out of 4 for UK
		SchoolsCount:  5,
		Source:        "placeholder",
	}, nil
}

// getGenericEducationData gets generic education data
func (es *EducationService) getGenericEducationData(lat, lng float64, city string) (*EducationData, error) {
	// Can use OpenStreetMap to find schools nearby
	// Or other international education databases
	
	return &EducationData{
		Score:         65.0,
		AverageRating: 7.0,
		SchoolsCount:  3,
		Source:        "generic",
	}, nil
}

// CalculateEducationScore calculates education score for a property
func (es *EducationService) CalculateEducationScore(property *models.Property) (float64, string, error) {
	eduData, err := es.GetEducationData(
		property.Latitude,
		property.Longitude,
		property.Country,
		property.City,
	)
	
	if err != nil {
		return 60.0, "{}", err
	}
	
	// Calculate score based on data
	score := eduData.AverageRating * 10 // Convert rating to 0-100 scale
	if score > 100 {
		score = 100
	}
	
	// Adjust based on number of schools
	if eduData.SchoolsCount > 5 {
		score += 10
	} else if eduData.SchoolsCount > 2 {
		score += 5
	}
	
	if score > 100 {
		score = 100
	}
	
	eduData.Score = score
	
	dataJSON, _ := json.Marshal(eduData)
	
	return score, string(dataJSON), nil
}

