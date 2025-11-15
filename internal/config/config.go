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
	Host     string `json:"host"`
	Path     string `json:"path,omitempty"`
	Username string `json:"username,omitempty"`
}

// NormalizeHost removes the protocol prefix from the host if present
// Jenkins hosts should never have a protocol prefix (http:// or https://)
// They should be stored as just the hostname (e.g., build.intuit.com)
func NormalizeHost(host string) string {
	// Remove http:// or https:// prefix if present
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	// Remove trailing slash if present
	host = strings.TrimSuffix(host, "/")
	return host
}

// NormalizePath removes leading and trailing slashes from path if present
// Paths should be stored without leading/trailing slashes (e.g., "jenkins" not "/jenkins/")
func NormalizePath(path string) string {
	// Remove leading and trailing slashes
	path = strings.Trim(path, "/")
	return path
}

// FormatHostURL returns the full HTTPS URL for the host with optional path
// Always uses HTTPS as required
func FormatHostURL(host, path string) string {
	// Normalize first to ensure no protocol prefix
	host = NormalizeHost(host)
	path = NormalizePath(path)
	
	url := "https://" + host
	if path != "" {
		url = url + "/" + path
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

// SaveConfig saves the host, path, and username to the config file
func SaveConfig(host, path, username string) error {
	// Normalize host to remove any protocol prefix
	host = NormalizeHost(host)
	// Normalize path to remove leading/trailing slashes
	path = NormalizePath(path)

	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDirPath := filepath.Dir(configPath)
	if err := os.MkdirAll(configDirPath, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	cfg := config{Host: host, Path: path, Username: username}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadConfig loads the host, path, and username from the config file
func LoadConfig() (string, string, string, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return "", "", "", err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", "", "", fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg.Host, cfg.Path, cfg.Username, nil
}

// SaveToken saves the token to the keyring
func SaveToken(host, token string) error {
	return keyring.Set(serviceName, host, token)
}

// LoadToken loads the token from the keyring
func LoadToken(host string) (string, error) {
	return keyring.Get(serviceName, host)
}
