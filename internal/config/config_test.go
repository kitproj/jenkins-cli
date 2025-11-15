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

// TestNormalizePath tests the path normalization function
func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "path without slashes",
			input:    "jenkins",
			expected: "jenkins",
		},
		{
			name:     "path with leading slash",
			input:    "/jenkins",
			expected: "jenkins",
		},
		{
			name:     "path with trailing slash",
			input:    "jenkins/",
			expected: "jenkins",
		},
		{
			name:     "path with both slashes",
			input:    "/jenkins/",
			expected: "jenkins",
		},
		{
			name:     "empty path",
			input:    "",
			expected: "",
		},
		{
			name:     "path with multiple segments",
			input:    "/ci/jenkins/",
			expected: "ci/jenkins",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizePath(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestFormatHostURL tests the host URL formatting function
func TestFormatHostURL(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		path     string
		expected string
	}{
		{
			name:     "host without protocol or path",
			host:     "build.intuit.com",
			path:     "",
			expected: "https://build.intuit.com",
		},
		{
			name:     "host with https protocol - should be normalized first",
			host:     "https://build.intuit.com",
			path:     "",
			expected: "https://build.intuit.com",
		},
		{
			name:     "host with http protocol - should convert to https",
			host:     "http://build.intuit.com",
			path:     "",
			expected: "https://build.intuit.com",
		},
		{
			name:     "localhost with port",
			host:     "localhost:8080",
			path:     "",
			expected: "https://localhost:8080",
		},
		{
			name:     "host with path",
			host:     "build.intuit.com",
			path:     "jenkins",
			expected: "https://build.intuit.com/jenkins",
		},
		{
			name:     "host with path having leading slash",
			host:     "build.intuit.com",
			path:     "/jenkins",
			expected: "https://build.intuit.com/jenkins",
		},
		{
			name:     "host with path having trailing slash",
			host:     "build.intuit.com",
			path:     "jenkins/",
			expected: "https://build.intuit.com/jenkins",
		},
		{
			name:     "host with path having both slashes",
			host:     "build.intuit.com",
			path:     "/jenkins/",
			expected: "https://build.intuit.com/jenkins",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatHostURL(tt.host, tt.path)
			if result != tt.expected {
				t.Errorf("FormatHostURL(%q, %q) = %q, want %q", tt.host, tt.path, result, tt.expected)
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
	testPath := "jenkins"
	testUsername := "testuser"

	// Test SaveConfig
	err := SaveConfig(testHost, testPath, testUsername)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Test LoadConfig
	retrievedHost, retrievedPath, retrievedUsername, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if retrievedHost != testHost {
		t.Errorf("Expected host %q, got %q", testHost, retrievedHost)
	}

	if retrievedPath != testPath {
		t.Errorf("Expected path %q, got %q", testPath, retrievedPath)
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
	err := SaveConfig(testHostWithProtocol, "", testUsername)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Test LoadConfig - should return host without protocol
	retrievedHost, _, _, err := LoadConfig()
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

	// Test SaveConfig without username or path
	err := SaveConfig(testHost, "", "")
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Test LoadConfig
	retrievedHost, retrievedPath, retrievedUsername, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if retrievedHost != testHost {
		t.Errorf("Expected host %q, got %q", testHost, retrievedHost)
	}

	if retrievedPath != "" {
		t.Errorf("Expected empty path, got %q", retrievedPath)
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
	_, _, _, err := LoadConfig()
	if err == nil {
		t.Error("Expected error when loading non-existent config, got nil")
	}
}
