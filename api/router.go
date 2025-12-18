package api

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	// Middleware
	router.Use(LoggerMiddleware())
	router.Use(CORSMiddleware())
	router.Use(RateLimitMiddleware())

	handler := NewHandler()

	// Health check endpoints (for monitoring and K8s probes)
	router.GET("/health", HealthHandler)
	router.GET("/readiness", ReadinessHandler)
	router.GET("/liveness", LivenessHandler)

	// Static files (frontend)
	router.Static("/web", "./web")
	router.StaticFile("/", "./web/index.html")
	router.StaticFile("/index.html", "./web/index.html")

	// Serve static assets with correct paths
	router.StaticFile("/style.css", "./web/style.css")
	router.StaticFile("/app.js", "./web/app.js")
	router.StaticFile("/app-leaflet.js", "./web/app-leaflet.js")

	// API routes
	api := router.Group("/api/v1")
	{
		api.GET("/heatmap", handler.GetHeatmapData)
		api.GET("/properties", handler.GetProperties)
		api.GET("/properties/:id", handler.GetPropertyDetails)
		api.GET("/stats", handler.GetStats)
		api.GET("/metrics", handler.GetMetrics)
		api.GET("/metrics/parser/:parser", handler.GetParserMetrics)
	}

	return router
}
