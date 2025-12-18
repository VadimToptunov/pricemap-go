package utils

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

// ProxyInfo represents a single proxy
type ProxyInfo struct {
	URL          string
	Protocol     string // "http", "https", "socks5"
	IsWorking    bool
	FailureCount int
	LastChecked  time.Time
	LastUsed     time.Time
}

// ProxyPool manages a pool of proxies
type ProxyPool struct {
	proxies     []*ProxyInfo
	currentIdx  int
	mu          sync.RWMutex
	maxFailures int
}

// NewProxyPool creates a new proxy pool
func NewProxyPool() *ProxyPool {
	return &ProxyPool{
		proxies:     make([]*ProxyInfo, 0),
		currentIdx:  0,
		maxFailures: 3,
	}
}

// AddProxy adds a proxy to the pool
func (pp *ProxyPool) AddProxy(proxyURL, protocol string) error {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	// Validate proxy URL
	parsed, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %w", err)
	}

	// Check that URL has a scheme and host
	if parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("invalid proxy URL: missing scheme or host")
	}

	proxyInfo := &ProxyInfo{
		URL:         proxyURL,
		Protocol:    protocol,
		IsWorking:   true,
		LastChecked: time.Now(),
	}

	pp.proxies = append(pp.proxies, proxyInfo)
	log.Printf("Added proxy: %s (%s)", proxyURL, protocol)

	return nil
}

// AddProxiesFromList adds multiple proxies from a list
func (pp *ProxyPool) AddProxiesFromList(proxies []string, protocol string) {
	for _, proxyURL := range proxies {
		if err := pp.AddProxy(proxyURL, protocol); err != nil {
			log.Printf("Failed to add proxy %s: %v", proxyURL, err)
		}
	}
}

// GetNextProxy returns the next working proxy in round-robin fashion
func (pp *ProxyPool) GetNextProxy() (*ProxyInfo, error) {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	if len(pp.proxies) == 0 {
		return nil, fmt.Errorf("no proxies available")
	}

	// Try to find a working proxy
	attempts := 0
	maxAttempts := len(pp.proxies)

	for attempts < maxAttempts {
		pp.currentIdx = (pp.currentIdx + 1) % len(pp.proxies)
		proxyInfo := pp.proxies[pp.currentIdx]

		if proxyInfo.IsWorking {
			proxyInfo.LastUsed = time.Now()
			return proxyInfo, nil
		}

		attempts++
	}

	return nil, fmt.Errorf("no working proxies available")
}

// GetRandomProxy returns a random working proxy
func (pp *ProxyPool) GetRandomProxy() (*ProxyInfo, error) {
	pp.mu.RLock()
	defer pp.mu.RUnlock()

	if len(pp.proxies) == 0 {
		return nil, fmt.Errorf("no proxies available")
	}

	// Get all working proxies
	workingProxies := make([]*ProxyInfo, 0)
	for _, p := range pp.proxies {
		if p.IsWorking {
			workingProxies = append(workingProxies, p)
		}
	}

	if len(workingProxies) == 0 {
		return nil, fmt.Errorf("no working proxies available")
	}

	// Return random working proxy
	idx := rand.Intn(len(workingProxies))
	proxy := workingProxies[idx]
	proxy.LastUsed = time.Now()

	return proxy, nil
}

// MarkProxyFailed marks a proxy as failed
func (pp *ProxyPool) MarkProxyFailed(proxyInfo *ProxyInfo) {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	proxyInfo.FailureCount++

	if proxyInfo.FailureCount >= pp.maxFailures {
		proxyInfo.IsWorking = false
		log.Printf("Proxy marked as not working after %d failures: %s", proxyInfo.FailureCount, proxyInfo.URL)
	}
}

// MarkProxyWorking marks a proxy as working
func (pp *ProxyPool) MarkProxyWorking(proxyInfo *ProxyInfo) {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	proxyInfo.IsWorking = true
	proxyInfo.FailureCount = 0
	proxyInfo.LastChecked = time.Now()
}

// GetWorkingProxiesCount returns the number of working proxies
func (pp *ProxyPool) GetWorkingProxiesCount() int {
	pp.mu.RLock()
	defer pp.mu.RUnlock()

	count := 0
	for _, p := range pp.proxies {
		if p.IsWorking {
			count++
		}
	}

	return count
}

// CreateHTTPClient creates an HTTP client with the given proxy
func (pp *ProxyPool) CreateHTTPClient(proxyInfo *ProxyInfo, timeout time.Duration) (*http.Client, error) {
	transport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}

	switch proxyInfo.Protocol {
	case "http", "https":
		proxyURL, err := url.Parse(proxyInfo.URL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)

	case "socks5":
		// Parse proxy URL to get host:port
		proxyURL, err := url.Parse(proxyInfo.URL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}

		dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, nil, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
		}

		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		}

	default:
		return nil, fmt.Errorf("unsupported proxy protocol: %s", proxyInfo.Protocol)
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}, nil
}

// CheckProxies tests all proxies and updates their status
func (pp *ProxyPool) CheckProxies(testURL string) {
	pp.mu.RLock()
	proxies := make([]*ProxyInfo, len(pp.proxies))
	copy(proxies, pp.proxies)
	pp.mu.RUnlock()

	log.Printf("Checking %d proxies...", len(proxies))

	for _, proxyInfo := range proxies {
		if err := pp.testProxy(proxyInfo, testURL); err != nil {
			pp.MarkProxyFailed(proxyInfo)
			log.Printf("Proxy check failed for %s: %v", proxyInfo.URL, err)
		} else {
			pp.MarkProxyWorking(proxyInfo)
			log.Printf("Proxy check successful for %s", proxyInfo.URL)
		}
	}

	log.Printf("Proxy check complete. Working proxies: %d/%d", pp.GetWorkingProxiesCount(), len(proxies))
}

func (pp *ProxyPool) testProxy(proxyInfo *ProxyInfo, testURL string) error {
	client, err := pp.CreateHTTPClient(proxyInfo, 10*time.Second)
	if err != nil {
		return err
	}

	resp, err := client.Get(testURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// RemoveFailedProxies removes proxies that have been marked as not working
func (pp *ProxyPool) RemoveFailedProxies() int {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	workingProxies := make([]*ProxyInfo, 0)
	removed := 0

	for _, p := range pp.proxies {
		if p.IsWorking {
			workingProxies = append(workingProxies, p)
		} else {
			removed++
		}
	}

	pp.proxies = workingProxies

	if removed > 0 {
		log.Printf("Removed %d failed proxies", removed)
	}

	return removed
}
