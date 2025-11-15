package config

import (
	"os"
	"testing"
)

// TestNormalizeHost tests the host normalization function
func TestNormalizeHost(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "host without protocol",
			input:    "build.intuit.com",
			expected: "build.intuit.com",
		},
		{
			name:     "host with https protocol",
			input:    "https://build.intuit.com",
			expected: "build.intuit.com",
		},
		{
			name:     "host with http protocol",
			input:    "http://build.intuit.com",
			expected: "build.intuit.com",
		},
		{
			name:     "host with trailing slash",
			input:    "build.intuit.com/",
			expected: "build.intuit.com",
		},
		{
			name:     "host with protocol and trailing slash",
			input:    "https://build.intuit.com/",
			expected: "build.intuit.com",
		},
		{
			name:     "localhost with port",
			input:    "localhost:8080",
			expected: "localhost:8080",
		},
		{
			name:     "localhost with protocol and port",
			input:    "http://localhost:8080",
			expected: "localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeHost(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeHost(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestFormatHostURL tests the host URL formatting function
func TestFormatHostURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "host without protocol",
			input:    "build.intuit.com",
			expected: "https://build.intuit.com",
		},
		{
			name:     "host with https protocol - should be normalized first",
			input:    "https://build.intuit.com",
			expected: "https://build.intuit.com",
		},
		{
			name:     "host with http protocol - should convert to https",
			input:    "http://build.intuit.com",
			expected: "https://build.intuit.com",
		},
		{
			name:     "localhost with port",
			input:    "localhost:8080",
			expected: "https://localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatHostURL(tt.input)
			if result != tt.expected {
				t.Errorf("FormatHostURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

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

// TestSaveConfigNormalizesHost tests that SaveConfig normalizes the host
func TestSaveConfigNormalizesHost(t *testing.T) {
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

	testHostWithProtocol := "https://jenkins.example.com"
	expectedHost := "jenkins.example.com"
	testUsername := "testuser"

	// Test SaveConfig with protocol
	err := SaveConfig(testHostWithProtocol, testUsername)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Test LoadConfig - should return host without protocol
	retrievedHost, _, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if retrievedHost != expectedHost {
		t.Errorf("Expected normalized host %q, got %q", expectedHost, retrievedHost)
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
