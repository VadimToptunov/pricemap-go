package parsers

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/proxy"

	"pricemap-go/config"
	"pricemap-go/models"
)

// Parser interface for all parsers
type Parser interface {
	Name() string
	Parse(ctx context.Context) ([]models.Property, error)
	GetBaseURL() string
}

// BaseParser contains common logic for all parsers
type BaseParser struct {
	client  *http.Client
	baseURL string
}

func NewBaseParser(baseURL string) *BaseParser {
	transport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}

	// Configure Tor proxy if enabled
	if config.AppConfig.UseTor {
		torProxy := fmt.Sprintf("%s:%s", config.AppConfig.TorProxyHost, config.AppConfig.TorProxyPort)

		// Create SOCKS5 dialer for Tor
		dialer, err := proxy.SOCKS5("tcp", torProxy, nil, proxy.Direct)
		if err != nil {
			// If Tor connection fails, log and continue without Tor
			fmt.Printf("Warning: Failed to connect to Tor proxy at %s: %v. Continuing without Tor.\n", torProxy, err)
		} else {
			// Use Tor dialer for all connections
			transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.Dial(network, addr)
			}
		}
	}

	return &BaseParser{
		client: &http.Client{
			Timeout:   time.Duration(config.AppConfig.RequestTimeout) * time.Second,
			Transport: transport,
		},
		baseURL: baseURL,
	}
}

func (bp *BaseParser) GetBaseURL() string {
	return bp.baseURL
}

func (bp *BaseParser) Fetch(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set realistic headers to avoid blocking
	userAgent := config.AppConfig.UserAgent
	if userAgent == "" {
		userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,application/json,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	// Note: Do NOT set Accept-Encoding manually. Go's HTTP client automatically handles
	// compression when DisableCompression is false. Manual setting disables auto-decompression.
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Cache-Control", "max-age=0")

	resp, err := bp.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}

	// Only accept HTTP 200 (OK) for GET requests
	// HTTP 201 (Created) is only valid for POST/PUT requests that create resources
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
