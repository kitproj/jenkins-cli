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

	testURL := "https://jenkins.example.com/ci"
	testUsername := "testuser"

	// Test SaveConfig
	err := SaveConfig(testURL, testUsername)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Test LoadConfig
	retrievedURL, retrievedUsername, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if retrievedURL != testURL {
		t.Errorf("Expected URL %q, got %q", testURL, retrievedURL)
	}

	if retrievedUsername != testUsername {
		t.Errorf("Expected username %q, got %q", testUsername, retrievedUsername)
	}
}

// TestSaveConfigRemovesTrailingSlash tests that SaveConfig removes trailing slashes
func TestSaveConfigRemovesTrailingSlash(t *testing.T) {
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

	testURLWithTrailingSlash := "https://jenkins.example.com/"
	expectedURL := "https://jenkins.example.com"
	testUsername := "testuser"

	// Test SaveConfig with trailing slash
	err := SaveConfig(testURLWithTrailingSlash, testUsername)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Test LoadConfig - should return URL without trailing slash
	retrievedURL, _, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if retrievedURL != expectedURL {
		t.Errorf("Expected URL without trailing slash %q, got %q", expectedURL, retrievedURL)
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

	testURL := "https://jenkins.example.com"

	// Test SaveConfig without username
	err := SaveConfig(testURL, "")
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Test LoadConfig
	retrievedURL, retrievedUsername, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if retrievedURL != testURL {
		t.Errorf("Expected URL %q, got %q", testURL, retrievedURL)
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
