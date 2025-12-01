package api

import (
	"net/http"
	
	"github.com/gin-gonic/gin"
	"pricemap-go/services"
)

var metricsService *services.MetricsService

func init() {
	metricsService = services.NewMetricsService()
}

// GetMetrics returns system metrics
func (h *Handler) GetMetrics(c *gin.Context) {
	stats := metricsService.GetStats()
	c.JSON(http.StatusOK, stats)
}

// GetParserMetrics returns metrics for a specific parser
func (h *Handler) GetParserMetrics(c *gin.Context) {
	parserName := c.Param("parser")
	stats := metricsService.GetParserStats(parserName)
	
	if stats == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Parser not found"})
		return
	}
	
	c.JSON(http.StatusOK, stats)
}

