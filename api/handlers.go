package api

import (
	"net/http"
	"strconv"

	"pricemap-go/database"
	"pricemap-go/models"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// GetHeatmapData returns data for heatmap
func (h *Handler) GetHeatmapData(c *gin.Context) {
	// Request parameters
	latMin, _ := strconv.ParseFloat(c.Query("lat_min"), 64)
	latMax, _ := strconv.ParseFloat(c.Query("lat_max"), 64)
	lngMin, _ := strconv.ParseFloat(c.Query("lng_min"), 64)
	lngMax, _ := strconv.ParseFloat(c.Query("lng_max"), 64)

	// Grid size for aggregation
	gridSize := 0.01 // ~1km

	var properties []models.Property
	query := database.DB.Where("is_active = ?", true).
		Where("latitude != 0 AND longitude != 0").
		Where("latitude IS NOT NULL AND longitude IS NOT NULL")

	// Apply bounds filter only if provided
	if latMin != 0 || latMax != 0 || lngMin != 0 || lngMax != 0 {
		query = query.Where("latitude >= ? AND latitude <= ?", latMin, latMax).
			Where("longitude >= ? AND longitude <= ?", lngMin, lngMax)
	}

	// Apply property filters
	if city := c.Query("city"); city != "" {
		query = query.Where("city = ?", city)
	}
	if propertyType := c.Query("type"); propertyType != "" {
		query = query.Where("type = ?", propertyType)
	}
	if priceMin := c.Query("price_min"); priceMin != "" {
		if min, err := strconv.ParseFloat(priceMin, 64); err == nil {
			query = query.Where("price >= ?", min)
		}
	}
	if priceMax := c.Query("price_max"); priceMax != "" {
		if max, err := strconv.ParseFloat(priceMax, 64); err == nil {
			query = query.Where("price <= ?", max)
		}
	}
	if roomsMin := c.Query("rooms_min"); roomsMin != "" {
		if min, err := strconv.Atoi(roomsMin); err == nil {
			query = query.Where("rooms >= ?", min)
		}
	}
	if roomsMax := c.Query("rooms_max"); roomsMax != "" {
		if max, err := strconv.Atoi(roomsMax); err == nil {
			query = query.Where("rooms <= ?", max)
		}
	}
	if bedroomsMin := c.Query("bedrooms_min"); bedroomsMin != "" {
		if min, err := strconv.Atoi(bedroomsMin); err == nil {
			query = query.Where("bedrooms >= ?", min)
		}
	}
	if bedroomsMax := c.Query("bedrooms_max"); bedroomsMax != "" {
		if max, err := strconv.Atoi(bedroomsMax); err == nil {
			query = query.Where("bedrooms <= ?", max)
		}
	}
	if bathroomsMin := c.Query("bathrooms_min"); bathroomsMin != "" {
		if min, err := strconv.ParseFloat(bathroomsMin, 64); err == nil {
			query = query.Where("bathrooms >= ?", min)
		}
	}
	if bathroomsMax := c.Query("bathrooms_max"); bathroomsMax != "" {
		if max, err := strconv.ParseFloat(bathroomsMax, 64); err == nil {
			query = query.Where("bathrooms <= ?", max)
		}
	}
	if areaMin := c.Query("area_min"); areaMin != "" {
		if min, err := strconv.ParseFloat(areaMin, 64); err == nil {
			query = query.Where("area >= ?", min)
		}
	}
	if areaMax := c.Query("area_max"); areaMax != "" {
		if max, err := strconv.ParseFloat(areaMax, 64); err == nil {
			query = query.Where("area <= ?", max)
		}
	}

	// Score filters (from PropertyFactors)
	if scoreMin := c.Query("score_min"); scoreMin != "" {
		if min, err := strconv.ParseFloat(scoreMin, 64); err == nil {
			query = query.Joins("JOIN property_factors ON property_factors.property_id = properties.id").
				Where("property_factors.overall_score >= ?", min)
		}
	}
	if crimeScoreMin := c.Query("crime_score_min"); crimeScoreMin != "" {
		if min, err := strconv.ParseFloat(crimeScoreMin, 64); err == nil {
			query = query.Joins("JOIN property_factors ON property_factors.property_id = properties.id").
				Where("property_factors.crime_score >= ?", min)
		}
	}
	if transportScoreMin := c.Query("transport_score_min"); transportScoreMin != "" {
		if min, err := strconv.ParseFloat(transportScoreMin, 64); err == nil {
			query = query.Joins("JOIN property_factors ON property_factors.property_id = properties.id").
				Where("property_factors.transport_score >= ?", min)
		}
	}
	if educationScoreMin := c.Query("education_score_min"); educationScoreMin != "" {
		if min, err := strconv.ParseFloat(educationScoreMin, 64); err == nil {
			query = query.Joins("JOIN property_factors ON property_factors.property_id = properties.id").
				Where("property_factors.education_score >= ?", min)
		}
	}

	query = query.Preload("Factors")

	if err := query.Find(&properties).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Aggregate data by grid
	heatmapData := aggregateToHeatmap(properties, gridSize)

	c.JSON(http.StatusOK, gin.H{
		"data":  heatmapData,
		"count": len(properties),
	})
}

// GetPropertyDetails returns detailed information about a property
func (h *Handler) GetPropertyDetails(c *gin.Context) {
	id := c.Param("id")

	var property models.Property
	if err := database.DB.Preload("Factors").First(&property, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
		return
	}

	c.JSON(http.StatusOK, property)
}

// GetProperties returns list of properties with filters
func (h *Handler) GetProperties(c *gin.Context) {
	var properties []models.Property
	query := database.DB.Where("is_active = ?", true)

	// Filters
	if city := c.Query("city"); city != "" {
		query = query.Where("city = ?", city)
	}
	if country := c.Query("country"); country != "" {
		query = query.Where("country = ?", country)
	}
	if propertyType := c.Query("type"); propertyType != "" {
		query = query.Where("type = ?", propertyType)
	}

	// Price
	if priceMin := c.Query("price_min"); priceMin != "" {
		if min, err := strconv.ParseFloat(priceMin, 64); err == nil {
			query = query.Where("price >= ?", min)
		}
	}
	if priceMax := c.Query("price_max"); priceMax != "" {
		if max, err := strconv.ParseFloat(priceMax, 64); err == nil {
			query = query.Where("price <= ?", max)
		}
	}

	// Rooms
	if roomsMin := c.Query("rooms_min"); roomsMin != "" {
		if min, err := strconv.Atoi(roomsMin); err == nil {
			query = query.Where("rooms >= ?", min)
		}
	}
	if roomsMax := c.Query("rooms_max"); roomsMax != "" {
		if max, err := strconv.Atoi(roomsMax); err == nil {
			query = query.Where("rooms <= ?", max)
		}
	}

	// Bedrooms
	if bedroomsMin := c.Query("bedrooms_min"); bedroomsMin != "" {
		if min, err := strconv.Atoi(bedroomsMin); err == nil {
			query = query.Where("bedrooms >= ?", min)
		}
	}
	if bedroomsMax := c.Query("bedrooms_max"); bedroomsMax != "" {
		if max, err := strconv.Atoi(bedroomsMax); err == nil {
			query = query.Where("bedrooms <= ?", max)
		}
	}

	// Bathrooms
	if bathroomsMin := c.Query("bathrooms_min"); bathroomsMin != "" {
		if min, err := strconv.ParseFloat(bathroomsMin, 64); err == nil {
			query = query.Where("bathrooms >= ?", min)
		}
	}
	if bathroomsMax := c.Query("bathrooms_max"); bathroomsMax != "" {
		if max, err := strconv.ParseFloat(bathroomsMax, 64); err == nil {
			query = query.Where("bathrooms <= ?", max)
		}
	}

	// Area
	if areaMin := c.Query("area_min"); areaMin != "" {
		if min, err := strconv.ParseFloat(areaMin, 64); err == nil {
			query = query.Where("area >= ?", min)
		}
	}
	if areaMax := c.Query("area_max"); areaMax != "" {
		if max, err := strconv.ParseFloat(areaMax, 64); err == nil {
			query = query.Where("area <= ?", max)
		}
	}

	// Score filters (from PropertyFactors)
	if scoreMin := c.Query("score_min"); scoreMin != "" {
		if min, err := strconv.ParseFloat(scoreMin, 64); err == nil {
			query = query.Joins("JOIN property_factors ON property_factors.property_id = properties.id").
				Where("property_factors.overall_score >= ?", min)
		}
	}
	if crimeScoreMin := c.Query("crime_score_min"); crimeScoreMin != "" {
		if min, err := strconv.ParseFloat(crimeScoreMin, 64); err == nil {
			query = query.Joins("JOIN property_factors ON property_factors.property_id = properties.id").
				Where("property_factors.crime_score >= ?", min)
		}
	}
	if transportScoreMin := c.Query("transport_score_min"); transportScoreMin != "" {
		if min, err := strconv.ParseFloat(transportScoreMin, 64); err == nil {
			query = query.Joins("JOIN property_factors ON property_factors.property_id = properties.id").
				Where("property_factors.transport_score >= ?", min)
		}
	}
	if educationScoreMin := c.Query("education_score_min"); educationScoreMin != "" {
		if min, err := strconv.ParseFloat(educationScoreMin, 64); err == nil {
			query = query.Joins("JOIN property_factors ON property_factors.property_id = properties.id").
				Where("property_factors.education_score >= ?", min)
		}
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset := (page - 1) * limit

	var total int64
	query.Model(&models.Property{}).Count(&total)

	if err := query.Preload("Factors").
		Offset(offset).
		Limit(limit).
		Find(&properties).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  properties,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetStats returns statistics
func (h *Handler) GetStats(c *gin.Context) {
	var stats struct {
		TotalProperties int64    `json:"total_properties"`
		AvgPrice        float64  `json:"avg_price"`
		Countries       []string `json:"countries"`
		Cities          []string `json:"cities"`
	}

	database.DB.Model(&models.Property{}).
		Where("is_active = ?", true).
		Count(&stats.TotalProperties)

	database.DB.Model(&models.Property{}).
		Where("is_active = ?", true).
		Select("AVG(price)").Scan(&stats.AvgPrice)

	database.DB.Model(&models.Property{}).
		Where("is_active = ?", true).
		Distinct("country").
		Pluck("country", &stats.Countries)

	database.DB.Model(&models.Property{}).
		Where("is_active = ?", true).
		Distinct("city").
		Pluck("city", &stats.Cities)

	c.JSON(http.StatusOK, stats)
}

// aggregateToHeatmap aggregates properties into points for heatmap
func aggregateToHeatmap(properties []models.Property, gridSize float64) []models.PriceHeatmapPoint {
	grid := make(map[string]*models.PriceHeatmapPoint)

	for _, prop := range properties {
		// Round coordinates to grid size
		lat := roundToGrid(prop.Latitude, gridSize)
		lng := roundToGrid(prop.Longitude, gridSize)

		key := formatGridKey(lat, lng)

		if point, exists := grid[key]; exists {
			point.Price += prop.Price
			point.Count++
			if prop.Factors.OverallScore > 0 {
				point.Score = (point.Score*float64(point.Count-1) + prop.Factors.OverallScore) / float64(point.Count)
			}
		} else {
			score := 0.0
			if prop.Factors.OverallScore > 0 {
				score = prop.Factors.OverallScore
			}
			grid[key] = &models.PriceHeatmapPoint{
				Latitude:  lat,
				Longitude: lng,
				Price:     prop.Price,
				Score:     score,
				Count:     1,
			}
		}
	}

	// Convert to array and calculate average prices
	result := make([]models.PriceHeatmapPoint, 0, len(grid))
	for _, point := range grid {
		if point.Count > 0 {
			point.Price = point.Price / float64(point.Count)
		}
		result = append(result, *point)
	}

	return result
}

func roundToGrid(value, gridSize float64) float64 {
	return float64(int(value/gridSize)) * gridSize
}

func formatGridKey(lat, lng float64) string {
	return strconv.FormatFloat(lat, 'f', 6, 64) + "," + strconv.FormatFloat(lng, 'f', 6, 64)
}
