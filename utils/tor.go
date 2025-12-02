package utils

import (
	"fmt"
	"net"
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

	// This would require making a request through Tor and checking the IP
	// For now, we'll leave this as a placeholder
	return "", fmt.Errorf("not implemented")
}
