package config

import (
	"log"
	"os"
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
		RequestTimeout: 30,
		
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

