package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"pricemap-go/config"
	"pricemap-go/database"
	"pricemap-go/services"
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

	// Create scraper service
	scraperService := services.NewScraperService()

	// Create context with cancellation capability
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		cancel()
	}()

	// Start parsing (concurrent mode with 3 workers for better performance)
	log.Println("Starting scraper...")

	// Use concurrent scraping for better performance
	// Workers=3 is a good balance between speed and resource usage
	if err := scraperService.ScrapeAllConcurrent(ctx, 3); err != nil {
		log.Printf("Scraping completed with errors: %v", err)
		// Don't fatal - some parsers may have succeeded
	} else {
		log.Println("Scraper completed successfully")
	}
}
