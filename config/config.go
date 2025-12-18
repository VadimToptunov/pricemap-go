package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Server
	ServerPort string

	// API Keys
	GoogleMapsAPIKey string
	OpenCageAPIKey   string // For geocoding

	// Scraping
	UserAgent      string
	RequestTimeout int // seconds

	// Tor Proxy (for bypassing blocks)
	UseTor             bool
	TorProxyHost       string
	TorProxyPort       string
	TorControlPort     string // For circuit rotation
	TorControlPassword string

	// Rate Limiting & Retry
	RateLimitDelay int // seconds between requests
	MaxRetries     int
	RetryDelay     int // seconds

	// Cron
	CronSchedule string
}

var AppConfig *Config

func Load() {
	// Load .env file if it exists
	_ = godotenv.Load()

	AppConfig = &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "pricemap"),

		ServerPort: getEnv("SERVER_PORT", "3000"),

		GoogleMapsAPIKey: getEnv("GOOGLE_MAPS_API_KEY", ""),
		OpenCageAPIKey:   getEnv("OPENCAGE_API_KEY", ""),

		UserAgent:      getEnv("USER_AGENT", "PriceMap-Go/1.0"),
		RequestTimeout: getEnvInt("REQUEST_TIMEOUT", 30),

		UseTor:             getEnv("USE_TOR", "false") == "true",
		TorProxyHost:       getEnv("TOR_PROXY_HOST", "127.0.0.1"),
		TorProxyPort:       getEnv("TOR_PROXY_PORT", "9050"),
		TorControlPort:     getEnv("TOR_CONTROL_PORT", "9051"),
		TorControlPassword: getEnv("TOR_CONTROL_PASSWORD", ""),

		RateLimitDelay: getEnvInt("RATE_LIMIT_DELAY", 3),
		MaxRetries:     getEnvInt("MAX_RETRIES", 3),
		RetryDelay:     getEnvInt("RETRY_DELAY", 5),

		CronSchedule: getEnv("CRON_SCHEDULE", "0 */6 * * *"), // Every 6 hours
	}

	log.Println("Configuration loaded successfully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
