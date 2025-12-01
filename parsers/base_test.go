package parsers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"pricemap-go/config"
)

func init() {
	// Load config for tests
	config.Load()
}

func TestBaseParser_Fetch(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	bp := NewBaseParser(server.URL)

	ctx := context.Background()
	body, err := bp.Fetch(ctx, server.URL)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	defer body.Close()

	// Read body
	buf := make([]byte, 100)
	n, err := body.Read(buf)
	if err != nil && err.Error() != "EOF" {
		t.Fatalf("Read() error = %v", err)
	}

	if string(buf[:n]) != "test response" {
		t.Errorf("Fetch() body = %v, want 'test response'", string(buf[:n]))
	}
}

func TestBaseParser_Fetch_Error(t *testing.T) {
	bp := NewBaseParser("http://invalid-url-that-does-not-exist.local")

	ctx := context.Background()
	_, err := bp.Fetch(ctx, "http://invalid-url-that-does-not-exist.local/test")
	if err == nil {
		t.Errorf("Fetch() should return error for invalid URL")
	}
}

func TestBaseParser_Fetch_404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	bp := NewBaseParser(server.URL)

	ctx := context.Background()
	_, err := bp.Fetch(ctx, server.URL)
	if err == nil {
		t.Errorf("Fetch() should return error for 404 status")
	}
}

