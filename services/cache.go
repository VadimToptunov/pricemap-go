package services

import (
	"sync"
	"time"
)

// CacheService provides in-memory caching
type CacheService struct {
	cache map[string]*CacheEntry
	mu    sync.RWMutex
	ttl   time.Duration
}

type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

func NewCacheService(ttl time.Duration) *CacheService {
	cs := &CacheService{
		cache: make(map[string]*CacheEntry),
		ttl:   ttl,
	}
	
	// Start cleanup goroutine
	go cs.cleanup()
	
	return cs
}

// Get retrieves a value from cache
func (cs *CacheService) Get(key string) (interface{}, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	
	entry, exists := cs.cache[key]
	if !exists {
		return nil, false
	}
	
	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}
	
	return entry.Data, true
}

// Set stores a value in cache
func (cs *CacheService) Set(key string, value interface{}) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	cs.cache[key] = &CacheEntry{
		Data:      value,
		ExpiresAt: time.Now().Add(cs.ttl),
	}
}

// Delete removes a key from cache
func (cs *CacheService) Delete(key string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	delete(cs.cache, key)
}

// Clear removes all entries from cache
func (cs *CacheService) Clear() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.cache = make(map[string]*CacheEntry)
}

// cleanup removes expired entries periodically
func (cs *CacheService) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		cs.mu.Lock()
		now := time.Now()
		for key, entry := range cs.cache {
			if now.After(entry.ExpiresAt) {
				delete(cs.cache, key)
			}
		}
		cs.mu.Unlock()
	}
}

// GetStats returns cache statistics
func (cs *CacheService) GetStats() map[string]interface{} {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	
	return map[string]interface{}{
		"size":      len(cs.cache),
		"ttl_hours": cs.ttl.Hours(),
	}
}

