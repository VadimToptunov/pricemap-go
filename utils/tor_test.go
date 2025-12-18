package utils

import (
	"testing"
	"pricemap-go/config"
)

func TestNewTorController(t *testing.T) {
	// Setup config
	config.AppConfig = &config.Config{
		TorProxyHost:   "localhost",
		TorControlPort: "9051",
	}
	
	tc := NewTorController()
	
	if tc == nil {
		t.Fatal("NewTorController returned nil")
	}
	
	if tc.controlAddr != "localhost:9051" {
		t.Errorf("Expected controlAddr 'localhost:9051', got '%s'", tc.controlAddr)
	}
}

func TestTorController_RotateCircuit(t *testing.T) {
	// Skip if Tor is not available
	t.Skip("Requires running Tor instance")
	
	config.AppConfig = &config.Config{
		TorProxyHost:   "localhost",
		TorControlPort: "9051",
		TorControlPassword: "",
	}
	
	tc := NewTorController()
	err := tc.RotateCircuit()
	
	if err != nil {
		t.Logf("Tor rotation failed (expected if Tor not running): %v", err)
	} else {
		t.Log("Tor circuit rotated successfully")
	}
}

