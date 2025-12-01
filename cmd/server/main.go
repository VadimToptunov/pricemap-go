package main

import (
	"log"
	
	"pricemap-go/api"
	"pricemap-go/config"
	"pricemap-go/database"
)

func main() {
	// Load configuration
	config.Load()
	
	// Connect to database
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()
	
	// Run migrations
	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	
	// Setup router
	router := api.SetupRouter()
	
	// Start server
	port := ":" + config.AppConfig.ServerPort
	log.Printf("Server starting on port %s", port)
	if err := router.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

