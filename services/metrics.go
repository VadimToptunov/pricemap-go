package services

import (
	"sync"
	"time"
)

// MetricsService tracks system metrics
type MetricsService struct {
	mu sync.RWMutex
	
	// Parser metrics
	ParserStats map[string]*ParserStats
	
	// General metrics
	TotalPropertiesParsed int64
	TotalPropertiesSaved   int64
	TotalErrors           int64
	StartTime             time.Time
}

type ParserStats struct {
	PropertiesParsed int64
	PropertiesSaved  int64
	Errors          int64
	LastRunTime     time.Time
	AverageRunTime  time.Duration
	RunCount        int64
}

func NewMetricsService() *MetricsService {
	return &MetricsService{
		ParserStats: make(map[string]*ParserStats),
		StartTime:   time.Now(),
	}
}

// RecordParserRun records a parser execution
func (ms *MetricsService) RecordParserRun(parserName string, propertiesParsed, propertiesSaved int64, errors int64, duration time.Duration) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	stats, exists := ms.ParserStats[parserName]
	if !exists {
		stats = &ParserStats{}
		ms.ParserStats[parserName] = stats
	}
	
	stats.PropertiesParsed += propertiesParsed
	stats.PropertiesSaved += propertiesSaved
	stats.Errors += errors
	stats.LastRunTime = time.Now()
	stats.RunCount++
	
	// Update average run time
	if stats.RunCount > 0 {
		stats.AverageRunTime = (stats.AverageRunTime*time.Duration(stats.RunCount-1) + duration) / time.Duration(stats.RunCount)
	}
	
	ms.TotalPropertiesParsed += propertiesParsed
	ms.TotalPropertiesSaved += propertiesSaved
	ms.TotalErrors += errors
}

// GetStats returns current metrics
func (ms *MetricsService) GetStats() map[string]interface{} {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	uptime := time.Since(ms.StartTime)
	
	return map[string]interface{}{
		"uptime_seconds":      uptime.Seconds(),
		"total_parsed":        ms.TotalPropertiesParsed,
		"total_saved":         ms.TotalPropertiesSaved,
		"total_errors":       ms.TotalErrors,
		"parser_stats":       ms.ParserStats,
		"properties_per_sec": float64(ms.TotalPropertiesParsed) / uptime.Seconds(),
	}
}

// GetParserStats returns stats for a specific parser
func (ms *MetricsService) GetParserStats(parserName string) *ParserStats {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	return ms.ParserStats[parserName]
}

