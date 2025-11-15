package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kitproj/jenkins-cli/internal/keyring"
)

const (
	serviceName = "jenkins-cli"
	configFile  = "config.json"
)

// config represents the jenkins-cli configuration
type config struct {
	URL      string `json:"url"`
	Username string `json:"username,omitempty"`
}

// NormalizeURL ensures the URL has the https:// protocol and removes trailing slashes
// Returns the normalized URL that should be stored and used for connections
func NormalizeURL(url string) string {
	// Remove trailing slashes
	url = strings.TrimRight(url, "/")
	
	// If no protocol is specified, add https://
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
	
	return url
}

// getConfigPath returns the path to the config file
func getConfigPath() (string, error) {
	configDirPath, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}

	configPath := filepath.Join(configDirPath, "jenkins-cli", configFile)
	return configPath, nil
}

// SaveConfig saves the URL and username to the config file
func SaveConfig(url, username string) error {
	// Normalize URL to ensure it has protocol and no trailing slashes
	url = NormalizeURL(url)

	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDirPath := filepath.Dir(configPath)
	if err := os.MkdirAll(configDirPath, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	cfg := config{URL: url, Username: username}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadConfig loads the URL and username from the config file
func LoadConfig() (string, string, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return "", "", err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", "", fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg.URL, cfg.Username, nil
}

// SaveToken saves the token to the keyring
func SaveToken(url, token string) error {
	return keyring.Set(serviceName, url, token)
}

// LoadToken loads the token from the keyring
func LoadToken(url string) (string, error) {
	return keyring.Get(serviceName, url)
}
