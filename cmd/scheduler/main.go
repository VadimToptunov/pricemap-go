package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	
	"github.com/robfig/cron/v3"
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
	
	// Create scheduler
	c := cron.New()
	
	// Add parsing task
	scraperService := services.NewScraperService()
	
	_, err := c.AddFunc(config.AppConfig.CronSchedule, func() {
		log.Println("Starting scheduled scraping...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		
		if err := scraperService.ScrapeAll(ctx); err != nil {
			log.Printf("Scheduled scraping failed: %v", err)
		} else {
			log.Println("Scheduled scraping completed")
		}
	})
	
	if err != nil {
		log.Fatalf("Failed to schedule task: %v", err)
	}
	
	// Start scheduler
	c.Start()
	log.Printf("Scheduler started with schedule: %s", config.AppConfig.CronSchedule)
	
	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	// Run first execution immediately
	log.Println("Running initial scrape...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	
	if err := scraperService.ScrapeAll(ctx); err != nil {
		log.Printf("Initial scraping failed: %v", err)
	}
	
	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down scheduler...")
	c.Stop()
}

