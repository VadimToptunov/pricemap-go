package utils

import (
	"testing"
)

func TestNewProxyPool(t *testing.T) {
	pp := NewProxyPool()
	
	if pp == nil {
		t.Fatal("NewProxyPool returned nil")
	}
	
	if len(pp.proxies) != 0 {
		t.Errorf("Expected empty proxy pool, got %d proxies", len(pp.proxies))
	}
}

func TestProxyPool_AddProxy(t *testing.T) {
	pp := NewProxyPool()
	
	err := pp.AddProxy("socks5://127.0.0.1:9050", "socks5")
	if err != nil {
		t.Fatalf("Failed to add proxy: %v", err)
	}
	
	if len(pp.proxies) != 1 {
		t.Errorf("Expected 1 proxy, got %d", len(pp.proxies))
	}
	
	if pp.proxies[0].URL != "socks5://127.0.0.1:9050" {
		t.Errorf("Unexpected proxy URL: %s", pp.proxies[0].URL)
	}
}

func TestProxyPool_AddProxy_Invalid(t *testing.T) {
	pp := NewProxyPool()
	
	err := pp.AddProxy("not a valid url", "socks5")
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}

func TestProxyPool_GetNextProxy(t *testing.T) {
	pp := NewProxyPool()
	
	// Empty pool
	_, err := pp.GetNextProxy()
	if err == nil {
		t.Error("Expected error for empty pool, got nil")
	}
	
	// Add proxies
	pp.AddProxy("socks5://127.0.0.1:9050", "socks5")
	pp.AddProxy("socks5://127.0.0.1:9051", "socks5")
	
	proxy1, err := pp.GetNextProxy()
	if err != nil {
		t.Fatalf("Failed to get proxy: %v", err)
	}
	
	proxy2, err := pp.GetNextProxy()
	if err != nil {
		t.Fatalf("Failed to get proxy: %v", err)
	}
	
	// Should get different proxies (round-robin)
	if proxy1.URL == proxy2.URL {
		t.Error("Expected different proxies, got same")
	}
}

func TestProxyPool_MarkProxyFailed(t *testing.T) {
	pp := NewProxyPool()
	pp.maxFailures = 2
	
	pp.AddProxy("socks5://127.0.0.1:9050", "socks5")
	proxy := pp.proxies[0]
	
	// Mark as failed twice
	pp.MarkProxyFailed(proxy)
	if !proxy.IsWorking {
		t.Error("Proxy should still be working after 1 failure")
	}
	
	pp.MarkProxyFailed(proxy)
	if proxy.IsWorking {
		t.Error("Proxy should be marked as not working after 2 failures")
	}
}

func TestProxyPool_MarkProxyWorking(t *testing.T) {
	pp := NewProxyPool()
	
	pp.AddProxy("socks5://127.0.0.1:9050", "socks5")
	proxy := pp.proxies[0]
	
	// Mark as failed then working
	proxy.IsWorking = false
	proxy.FailureCount = 5
	
	pp.MarkProxyWorking(proxy)
	
	if !proxy.IsWorking {
		t.Error("Proxy should be marked as working")
	}
	
	if proxy.FailureCount != 0 {
		t.Errorf("Expected failure count 0, got %d", proxy.FailureCount)
	}
}

func TestProxyPool_GetWorkingProxiesCount(t *testing.T) {
	pp := NewProxyPool()
	
	pp.AddProxy("socks5://127.0.0.1:9050", "socks5")
	pp.AddProxy("socks5://127.0.0.1:9051", "socks5")
	pp.AddProxy("socks5://127.0.0.1:9052", "socks5")
	
	if pp.GetWorkingProxiesCount() != 3 {
		t.Errorf("Expected 3 working proxies, got %d", pp.GetWorkingProxiesCount())
	}
	
	// Mark one as failed
	pp.proxies[0].IsWorking = false
	
	if pp.GetWorkingProxiesCount() != 2 {
		t.Errorf("Expected 2 working proxies, got %d", pp.GetWorkingProxiesCount())
	}
}

func TestProxyPool_RemoveFailedProxies(t *testing.T) {
	pp := NewProxyPool()
	
	pp.AddProxy("socks5://127.0.0.1:9050", "socks5")
	pp.AddProxy("socks5://127.0.0.1:9051", "socks5")
	pp.AddProxy("socks5://127.0.0.1:9052", "socks5")
	
	// Mark two as failed
	pp.proxies[0].IsWorking = false
	pp.proxies[2].IsWorking = false
	
	removed := pp.RemoveFailedProxies()
	
	if removed != 2 {
		t.Errorf("Expected 2 removed proxies, got %d", removed)
	}
	
	if len(pp.proxies) != 1 {
		t.Errorf("Expected 1 proxy remaining, got %d", len(pp.proxies))
	}
}

