package utils

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/proxy"

	"pricemap-go/config"
)

// TorControl provides interface to Tor control port for circuit rotation
type TorControl struct {
	controlPort string
}

// NewTorControl creates a new Tor control interface
func NewTorControl() *TorControl {
	return &TorControl{
		controlPort: fmt.Sprintf("%s:%s", config.AppConfig.TorProxyHost, config.AppConfig.TorControlPort),
	}
}

// RenewCircuit sends NEWNYM command to Tor to get a new circuit
// This changes the IP address that Tor uses
func (tc *TorControl) RenewCircuit() error {
	if !config.AppConfig.UseTor {
		return fmt.Errorf("Tor is not enabled")
	}

	conn, err := net.Dial("tcp", tc.controlPort)
	if err != nil {
		return fmt.Errorf("failed to connect to Tor control port: %w", err)
	}
	defer conn.Close()

	// Send authentication (if password is set, use it; otherwise use empty auth)
	// For most default Tor installations, no password is needed
	authCmd := "AUTHENTICATE\n"
	_, err = conn.Write([]byte(authCmd))
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	// Read response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("failed to read auth response: %w", err)
	}

	response := string(buf[:n])
	if response[:3] != "250" {
		// Try with password if available
		// For now, we'll assume no password is needed
	}

	// Send NEWNYM command to get new circuit
	newNymCmd := "SIGNAL NEWNYM\n"
	_, err = conn.Write([]byte(newNymCmd))
	if err != nil {
		return fmt.Errorf("failed to send NEWNYM: %w", err)
	}

	// Read response
	n, err = conn.Read(buf)
	if err != nil {
		return fmt.Errorf("failed to read NEWNYM response: %w", err)
	}

	response = string(buf[:n])
	if response[:3] != "250" {
		return fmt.Errorf("Tor did not accept NEWNYM command: %s", response)
	}

	return nil
}

// GetCurrentIP returns the current IP address through Tor
func (tc *TorControl) GetCurrentIP() (string, error) {
	if !config.AppConfig.UseTor {
		return "", fmt.Errorf("Tor is not enabled")
	}

	// Create HTTP client that uses Tor SOCKS proxy
	torProxy := fmt.Sprintf("%s:%s", config.AppConfig.TorProxyHost, config.AppConfig.TorProxyPort)
	dialer, err := proxy.SOCKS5("tcp", torProxy, nil, proxy.Direct)
	if err != nil {
		return "", fmt.Errorf("failed to create Tor dialer: %w", err)
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	// Use a reliable IP checking service
	// Try multiple services in case one is down
	ipServices := []string{
		"https://api.ipify.org",
		"https://icanhazip.com",
		"https://ifconfig.me/ip",
	}

	var lastErr error
	for _, serviceURL := range ipServices {
		req, err := http.NewRequestWithContext(context.Background(), "GET", serviceURL, nil)
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}

		req.Header.Set("User-Agent", "Mozilla/5.0")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to make request to %s: %w", serviceURL, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, serviceURL)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("failed to read response from %s: %w", serviceURL, err)
			continue
		}

		ip := string(body)
		// Clean up the IP (remove whitespace, newlines, etc.)
		ip = strings.TrimSpace(ip)

		// Validate that it looks like an IP address
		if net.ParseIP(ip) != nil {
			return ip, nil
		}

		lastErr = fmt.Errorf("invalid IP address received: %s", ip)
	}

	return "", fmt.Errorf("failed to get IP from any service: %w", lastErr)
}
