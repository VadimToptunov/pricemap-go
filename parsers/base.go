package parsers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
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
	return &BaseParser{
		client: &http.Client{
			Timeout: time.Duration(config.AppConfig.RequestTimeout) * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
			},
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
	
	req.Header.Set("User-Agent", config.AppConfig.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	
	resp, err := bp.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	return resp.Body, nil
}

