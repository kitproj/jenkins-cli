package config

import (
	"os"
	"testing"
)

// TestSaveLoadConfig tests basic save and load operations
func TestSaveLoadConfig(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the config directory
	origConfigDir := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() {
		if origConfigDir != "" {
			os.Setenv("XDG_CONFIG_HOME", origConfigDir)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	testHost := "jenkins.example.com"
	testUsername := "testuser"

	// Test SaveConfig
	err := SaveConfig(testHost, testUsername)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Test LoadConfig
	retrievedHost, retrievedUsername, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if retrievedHost != testHost {
		t.Errorf("Expected host %q, got %q", testHost, retrievedHost)
	}

	if retrievedUsername != testUsername {
		t.Errorf("Expected username %q, got %q", testUsername, retrievedUsername)
	}
}

// TestSaveLoadConfigWithoutUsername tests save and load without username
func TestSaveLoadConfigWithoutUsername(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the config directory
	origConfigDir := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() {
		if origConfigDir != "" {
			os.Setenv("XDG_CONFIG_HOME", origConfigDir)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	testHost := "jenkins.example.com"

	// Test SaveConfig without username
	err := SaveConfig(testHost, "")
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Test LoadConfig
	retrievedHost, retrievedUsername, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if retrievedHost != testHost {
		t.Errorf("Expected host %q, got %q", testHost, retrievedHost)
	}

	if retrievedUsername != "" {
		t.Errorf("Expected empty username, got %q", retrievedUsername)
	}
}

// TestLoadConfigNotFound tests error handling when config doesn't exist
func TestLoadConfigNotFound(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the config directory
	origConfigDir := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() {
		if origConfigDir != "" {
			os.Setenv("XDG_CONFIG_HOME", origConfigDir)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	// Try to load from non-existent config
	_, _, err := LoadConfig()
	if err == nil {
		t.Error("Expected error when loading non-existent config, got nil")
	}
}
