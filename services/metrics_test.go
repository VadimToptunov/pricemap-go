package services

import (
	"testing"
	"time"
)

func TestMetricsService_RecordParserRun(t *testing.T) {
	ms := NewMetricsService()

	parserName := "test_parser"
	propertiesParsed := int64(100)
	propertiesSaved := int64(95)
	errors := int64(5)
	duration := 10 * time.Second

	ms.RecordParserRun(parserName, propertiesParsed, propertiesSaved, errors, duration)

	stats := ms.GetParserStats(parserName)
	if stats == nil {
		t.Fatalf("GetParserStats() returned nil")
	}

	if stats.PropertiesParsed != propertiesParsed {
		t.Errorf("PropertiesParsed = %v, want %v", stats.PropertiesParsed, propertiesParsed)
	}

	if stats.PropertiesSaved != propertiesSaved {
		t.Errorf("PropertiesSaved = %v, want %v", stats.PropertiesSaved, propertiesSaved)
	}

	if stats.Errors != errors {
		t.Errorf("Errors = %v, want %v", stats.Errors, errors)
	}

	if stats.RunCount != 1 {
		t.Errorf("RunCount = %v, want 1", stats.RunCount)
	}
}

func TestMetricsService_MultipleRuns(t *testing.T) {
	ms := NewMetricsService()

	parserName := "test_parser"

	// Record multiple runs
	ms.RecordParserRun(parserName, 100, 95, 5, 10*time.Second)
	ms.RecordParserRun(parserName, 200, 190, 10, 20*time.Second)
	ms.RecordParserRun(parserName, 150, 145, 5, 15*time.Second)

	stats := ms.GetParserStats(parserName)
	if stats.PropertiesParsed != 450 {
		t.Errorf("PropertiesParsed = %v, want 450", stats.PropertiesParsed)
	}

	if stats.RunCount != 3 {
		t.Errorf("RunCount = %v, want 3", stats.RunCount)
	}

	// Average run time should be calculated
	expectedAvg := (10*time.Second + 20*time.Second + 15*time.Second) / 3
	if stats.AverageRunTime != expectedAvg {
		t.Errorf("AverageRunTime = %v, want %v", stats.AverageRunTime, expectedAvg)
	}
}

func TestMetricsService_GetStats(t *testing.T) {
	ms := NewMetricsService()

	ms.RecordParserRun("parser1", 100, 95, 5, 10*time.Second)
	ms.RecordParserRun("parser2", 200, 190, 10, 20*time.Second)

	stats := ms.GetStats()

	if stats["total_parsed"].(int64) != 300 {
		t.Errorf("total_parsed = %v, want 300", stats["total_parsed"])
	}

	if stats["total_saved"].(int64) != 285 {
		t.Errorf("total_saved = %v, want 285", stats["total_saved"])
	}

	if stats["total_errors"].(int64) != 15 {
		t.Errorf("total_errors = %v, want 15", stats["total_errors"])
	}

	if stats["uptime_seconds"].(float64) <= 0 {
		t.Errorf("uptime_seconds should be positive")
	}
}

