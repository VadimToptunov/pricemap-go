package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"pricemap-go/config"
	"pricemap-go/database"
)

var startTime = time.Now()

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Uptime    string                 `json:"uptime"`
	Version   string                 `json:"version"`
	Checks    map[string]HealthCheck `json:"checks"`
}

// HealthCheck represents individual component health
type HealthCheck struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// HealthHandler handles health check requests
func HealthHandler(c *gin.Context) {
	checks := make(map[string]HealthCheck)

	// Check database
	dbCheck := checkDatabase()
	checks["database"] = dbCheck

	// Check Tor (if enabled)
	if config.AppConfig.UseTor {
		torCheck := checkTor()
		checks["tor"] = torCheck
	}

	// Determine overall status
	overallStatus := "healthy"
	for _, check := range checks {
		if check.Status != "healthy" {
			overallStatus = "degraded"
			break
		}
	}

	response := HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now().Format(time.RFC3339),
		Uptime:    time.Since(startTime).String(),
		Version:   "2.0",
		Checks:    checks,
	}

	statusCode := http.StatusOK
	if overallStatus != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

// ReadinessHandler handles readiness probe (for Kubernetes)
func ReadinessHandler(c *gin.Context) {
	// Check if database is accessible
	dbCheck := checkDatabase()

	if dbCheck.Status == "healthy" {
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"reason": dbCheck.Message,
		})
	}
}

// LivenessHandler handles liveness probe (for Kubernetes)
func LivenessHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}

// checkDatabase checks if database connection is healthy
func checkDatabase() HealthCheck {
	sqlDB, err := database.DB.DB()
	if err != nil {
		return HealthCheck{
			Status:  "unhealthy",
			Message: "failed to get database connection: " + err.Error(),
		}
	}

	if err := sqlDB.Ping(); err != nil {
		return HealthCheck{
			Status:  "unhealthy",
			Message: "database ping failed: " + err.Error(),
		}
	}

	return HealthCheck{
		Status:  "healthy",
		Message: "database connection ok",
	}
}

// checkTor checks if Tor proxy is accessible (basic check)
func checkTor() HealthCheck {
	// Simple check - just verify config is set
	// Full Tor connectivity check would require actual network call
	if config.AppConfig.TorProxyHost != "" && config.AppConfig.TorProxyPort != "" {
		return HealthCheck{
			Status:  "healthy",
			Message: "tor configured",
		}
	}

	return HealthCheck{
		Status:  "unknown",
		Message: "tor not configured",
	}
}

