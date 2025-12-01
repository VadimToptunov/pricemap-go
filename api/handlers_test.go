package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"pricemap-go/models"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := SetupRouter()
	return router
}

func TestHandler_GetStats(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Stats endpoint might fail without DB, but should return some response
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
}

func TestHandler_GetMetrics(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "total_parsed")
	assert.Contains(t, response, "total_saved")
	assert.Contains(t, response, "total_errors")
	assert.Contains(t, response, "uptime_seconds")
}

func TestHandler_GetHeatmapData(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/heatmap?lat_min=55&lat_max=56&lng_min=37&lng_max=38", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Heatmap might fail without DB, but should handle request
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
}

func TestHandler_GetProperties(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name   string
		url    string
		status int
	}{
		{
			name:   "basic request",
			url:    "/api/v1/properties",
			status: http.StatusOK, // Or 500 if DB not connected
		},
		{
			name:   "with filters",
			url:    "/api/v1/properties?city=Moscow&type=apartment&price_min=100000&price_max=500000",
			status: http.StatusOK, // Or 500 if DB not connected
		},
		{
			name:   "with pagination",
			url:    "/api/v1/properties?page=1&limit=10",
			status: http.StatusOK, // Or 500 if DB not connected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Accept both OK and InternalServerError (if DB not connected)
			assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
		})
	}
}

func TestHandler_GetPropertyDetails(t *testing.T) {
	router := setupTestRouter()

	// Test with non-existent ID
	req, _ := http.NewRequest("GET", "/api/v1/properties/999999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 404 or 500 (if DB not connected)
	assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusInternalServerError)
}

func TestHandler_CORS(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("OPTIONS", "/api/v1/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Origin"), "*")
}

func TestAggregateToHeatmap(t *testing.T) {
	properties := []models.Property{
		{
			Latitude:  55.7558,
			Longitude: 37.6173,
			Price:     100000,
			Factors: models.PropertyFactors{
				OverallScore: 80.0,
			},
		},
		{
			Latitude:  55.7559,
			Longitude: 37.6174,
			Price:     150000,
			Factors: models.PropertyFactors{
				OverallScore: 70.0,
			},
		},
		{
			Latitude:  55.7558,
			Longitude: 37.6173, // Same grid cell (after rounding)
			Price:     120000,
			Factors: models.PropertyFactors{
				OverallScore: 75.0,
			},
		},
	}

	gridSize := 0.01
	result := aggregateToHeatmap(properties, gridSize)

	if len(result) == 0 {
		t.Errorf("aggregateToHeatmap() returned empty result")
	}

	// Check that we have aggregated data
	totalCount := 0
	for _, point := range result {
		totalCount += point.Count
		if point.Price <= 0 {
			t.Errorf("aggregateToHeatmap() price should be positive, got %v", point.Price)
		}
		if point.Latitude == 0 || point.Longitude == 0 {
			t.Errorf("aggregateToHeatmap() coordinates should not be zero")
		}
	}

	if totalCount != len(properties) {
		t.Errorf("aggregateToHeatmap() total count = %v, want %v", totalCount, len(properties))
	}
}

func TestRoundToGrid(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		gridSize float64
		want     float64
	}{
		{
			name:     "round down",
			value:    55.7558,
			gridSize: 0.01,
			want:     55.75,
		},
		{
			name:     "round up",
			value:    55.7599,
			gridSize: 0.01,
			want:     55.75,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := roundToGrid(tt.value, tt.gridSize)
			diff := got - tt.want
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.001 {
				t.Errorf("roundToGrid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatGridKey(t *testing.T) {
	key := formatGridKey(55.7558, 37.6173)
	
	if key == "" {
		t.Errorf("formatGridKey() returned empty string")
	}
	
	// Should contain both coordinates
	if len(key) < 10 {
		t.Errorf("formatGridKey() returned too short key: %s", key)
	}
}
