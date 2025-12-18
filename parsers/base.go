package parsers

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/proxy"

	"pricemap-go/config"
	"pricemap-go/models"
	"pricemap-go/utils"
)

// Parser interface for all parsers
type Parser interface {
	Name() string
	Parse(ctx context.Context) ([]models.Property, error)
	GetBaseURL() string
}

// BaseParser contains common logic for all parsers
type BaseParser struct {
	client        *http.Client
	baseURL       string
	torController *utils.TorController
	requestCount  int
	mu            sync.Mutex // Protects requestCount for concurrent access
}

func NewBaseParser(baseURL string) *BaseParser {
	transport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}

	var torController *utils.TorController

	// Configure Tor proxy if enabled
	if config.AppConfig.UseTor {
		torProxy := fmt.Sprintf("%s:%s", config.AppConfig.TorProxyHost, config.AppConfig.TorProxyPort)

		// Create SOCKS5 dialer for Tor
		dialer, err := proxy.SOCKS5("tcp", torProxy, nil, proxy.Direct)
		if err != nil {
			// If Tor connection fails, log and continue without Tor
			log.Printf("Warning: Failed to connect to Tor proxy at %s: %v. Continuing without Tor.\n", torProxy, err)
		} else {
			// Use Tor dialer for all connections
			transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.Dial(network, addr)
			}

			torController = utils.NewTorController()
			log.Println("Tor proxy configured successfully")
		}
	}

	return &BaseParser{
		client: &http.Client{
			Timeout:   time.Duration(config.AppConfig.RequestTimeout) * time.Second,
			Transport: transport,
		},
		baseURL:       baseURL,
		torController: torController,
		requestCount:  0,
	}
}

func (bp *BaseParser) GetBaseURL() string {
	return bp.baseURL
}

func (bp *BaseParser) Fetch(ctx context.Context, url string) (io.ReadCloser, error) {
	var lastErr error
	maxRetries := config.AppConfig.MaxRetries

	// Thread-safe access to requestCount
	bp.mu.Lock()
	currentCount := bp.requestCount

	// Rate limiting with random delay (before the request)
	if currentCount > 0 {
		bp.mu.Unlock()
		delay := utils.GetRandomDelay(config.AppConfig.RateLimitDelay, config.AppConfig.RateLimitDelay+2)
		time.Sleep(delay)
		bp.mu.Lock()
	}

	// Rotate Tor circuit every 10 requests to avoid tracking
	if bp.torController != nil && currentCount > 0 && currentCount%10 == 0 {
		bp.mu.Unlock()
		log.Println("Rotating Tor circuit...")
		if err := bp.torController.RotateCircuit(); err != nil {
			log.Printf("Failed to rotate Tor circuit: %v", err)
		}
		bp.mu.Lock()
	}

	// Increment request count ONCE per Fetch call (outside retry loop)
	bp.requestCount++
	bp.mu.Unlock()

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Attempt to fetch
		body, err := bp.fetchWithRetry(ctx, url, attempt)
		if err == nil {
			return body, nil
		}

		lastErr = err

		// Don't retry on context cancellation
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Exponential backoff
		if attempt < maxRetries {
			backoffDelay := time.Duration(math.Pow(2, float64(attempt))) * time.Second * time.Duration(config.AppConfig.RetryDelay)
			log.Printf("Attempt %d failed: %v. Retrying in %v...", attempt+1, err, backoffDelay)

			// Rotate Tor circuit on retry if available
			if bp.torController != nil && attempt > 0 {
				if err := bp.torController.RotateCircuit(); err != nil {
					log.Printf("Failed to rotate Tor circuit: %v", err)
				}
			}

			time.Sleep(backoffDelay)
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries+1, lastErr)
}

func (bp *BaseParser) fetchWithRetry(ctx context.Context, url string, attempt int) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Use random user agent
	userAgent := utils.GetRandomUserAgent()
	if config.AppConfig.UserAgent != "" && config.AppConfig.UserAgent != "PriceMap-Go/1.0" {
		userAgent = config.AppConfig.UserAgent
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,application/json,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,ru;q=0.8")
	// Note: Do NOT set Accept-Encoding manually. Go's HTTP client automatically handles
	// compression when DisableCompression is false. Manual setting disables auto-decompression.
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("DNT", "1")

	resp, err := bp.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}

	// Accept 200 OK
	if resp.StatusCode == http.StatusOK {
		return resp.Body, nil
	}

	// Handle rate limiting (429) or server errors (5xx) with retry
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		resp.Body.Close()
		return nil, fmt.Errorf("server returned status %d (will retry)", resp.StatusCode)
	}

	// Other errors are final
	resp.Body.Close()
	return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
}
