package utils

import (
	"testing"
	"time"
)

func TestGetRandomUserAgent(t *testing.T) {
	ua := GetRandomUserAgent()
	
	if ua == "" {
		t.Error("GetRandomUserAgent returned empty string")
	}
	
	// Check that it's one of our user agents
	found := false
	for _, knownUA := range userAgents {
		if ua == knownUA {
			found = true
			break
		}
	}
	
	if !found {
		t.Errorf("GetRandomUserAgent returned unexpected UA: %s", ua)
	}
}

func TestGetRandomUserAgent_Different(t *testing.T) {
	// Test that we get different user agents
	uas := make(map[string]bool)
	
	for i := 0; i < 100; i++ {
		ua := GetRandomUserAgent()
		uas[ua] = true
	}
	
	// Should have at least 2 different user agents in 100 tries
	if len(uas) < 2 {
		t.Error("GetRandomUserAgent returns same UA too often")
	}
}

func TestGetRandomDelay(t *testing.T) {
	minSec := 1
	maxSec := 3
	
	delay := GetRandomDelay(minSec, maxSec)
	
	if delay < time.Second {
		t.Errorf("Delay too short: %v", delay)
	}
	
	if delay > time.Duration(maxSec+1)*time.Second {
		t.Errorf("Delay too long: %v", delay)
	}
}

func TestGetRandomDelay_SameMinMax(t *testing.T) {
	minSec := 2
	maxSec := 2
	
	delay := GetRandomDelay(minSec, maxSec)
	
	if delay != 2*time.Second {
		t.Errorf("Expected 2s delay, got %v", delay)
	}
}

