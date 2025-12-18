package utils

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"pricemap-go/config"
)

// TorController handles Tor circuit rotation
type TorController struct {
	controlAddr string
	password    string
}

// NewTorController creates a new Tor controller
func NewTorController() *TorController {
	return &TorController{
		controlAddr: fmt.Sprintf("%s:%s", config.AppConfig.TorProxyHost, config.AppConfig.TorControlPort),
		password:    config.AppConfig.TorControlPassword,
	}
}

// RotateCircuit requests a new Tor circuit (changes IP)
func (tc *TorController) RotateCircuit() error {
	conn, err := net.DialTimeout("tcp", tc.controlAddr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to Tor control port: %w", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Authenticate if password is set
	if tc.password != "" {
		if _, err := fmt.Fprintf(conn, "AUTHENTICATE \"%s\"\r\n", tc.password); err != nil {
			return fmt.Errorf("failed to send auth command: %w", err)
		}

		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read auth response: %w", err)
		}

		if !strings.HasPrefix(response, "250") {
			return fmt.Errorf("authentication failed: %s", response)
		}
	} else {
		// Try without password
		if _, err := fmt.Fprintf(conn, "AUTHENTICATE\r\n"); err != nil {
			return fmt.Errorf("failed to send auth command: %w", err)
		}

		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read auth response: %w", err)
		}

		if !strings.HasPrefix(response, "250") {
			return fmt.Errorf("authentication failed: %s", response)
		}
	}

	// Send NEWNYM signal to get new circuit
	if _, err := fmt.Fprintf(conn, "SIGNAL NEWNYM\r\n"); err != nil {
		return fmt.Errorf("failed to send NEWNYM signal: %w", err)
	}

	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read NEWNYM response: %w", err)
	}

	if !strings.HasPrefix(response, "250") {
		return fmt.Errorf("NEWNYM signal failed: %s", response)
	}

	log.Println("Tor circuit rotated successfully")

	// Wait a bit for new circuit to be established
	time.Sleep(2 * time.Second)

	return nil
}

// GetCircuitStatus returns the current Tor circuit status (for debugging)
func (tc *TorController) GetCircuitStatus() (string, error) {
	conn, err := net.DialTimeout("tcp", tc.controlAddr, 10*time.Second)
	if err != nil {
		return "", fmt.Errorf("failed to connect to Tor control port: %w", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Authenticate
	if tc.password != "" {
		if _, err := fmt.Fprintf(conn, "AUTHENTICATE \"%s\"\r\n", tc.password); err != nil {
			return "", fmt.Errorf("failed to send auth command: %w", err)
		}
	} else {
		if _, err := fmt.Fprintf(conn, "AUTHENTICATE\r\n"); err != nil {
			return "", fmt.Errorf("failed to send auth command: %w", err)
		}
	}

	response, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read auth response: %w", err)
	}

	if !strings.HasPrefix(response, "250") {
		return "", fmt.Errorf("authentication failed: %s", response)
	}

	// Get circuit status
	if _, err := fmt.Fprintf(conn, "GETINFO circuit-status\r\n"); err != nil {
		return "", fmt.Errorf("failed to send GETINFO command: %w", err)
	}

	status, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read circuit status: %w", err)
	}

	return status, nil
}
