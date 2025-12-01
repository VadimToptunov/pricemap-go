package services

import (
	"testing"
	"time"
)

func TestCacheService_GetSet(t *testing.T) {
	cs := NewCacheService(1 * time.Hour)

	// Test Set and Get
	key := "test_key"
	value := "test_value"

	cs.Set(key, value)

	got, exists := cs.Get(key)
	if !exists {
		t.Errorf("Get() exists = false, want true")
	}
	if got != value {
		t.Errorf("Get() = %v, want %v", got, value)
	}
}

func TestCacheService_Expiration(t *testing.T) {
	cs := NewCacheService(100 * time.Millisecond)

	key := "test_key"
	value := "test_value"

	cs.Set(key, value)

	// Should exist immediately
	_, exists := cs.Get(key)
	if !exists {
		t.Errorf("Get() immediately after Set should return true")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should not exist after expiration
	_, exists = cs.Get(key)
	if exists {
		t.Errorf("Get() after expiration should return false")
	}
}

func TestCacheService_Delete(t *testing.T) {
	cs := NewCacheService(1 * time.Hour)

	key := "test_key"
	value := "test_value"

	cs.Set(key, value)
	cs.Delete(key)

	_, exists := cs.Get(key)
	if exists {
		t.Errorf("Get() after Delete should return false")
	}
}

func TestCacheService_Clear(t *testing.T) {
	cs := NewCacheService(1 * time.Hour)

	cs.Set("key1", "value1")
	cs.Set("key2", "value2")
	cs.Set("key3", "value3")

	cs.Clear()

	stats := cs.GetStats()
	size := stats["size"].(int)
	if size != 0 {
		t.Errorf("GetStats() size after Clear = %v, want 0", size)
	}
}

func TestCacheService_GetStats(t *testing.T) {
	cs := NewCacheService(1 * time.Hour)

	cs.Set("key1", "value1")
	cs.Set("key2", "value2")

	stats := cs.GetStats()
	size := stats["size"].(int)
	if size != 2 {
		t.Errorf("GetStats() size = %v, want 2", size)
	}
}

