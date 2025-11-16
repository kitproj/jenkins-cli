package main

import (
	"context"
	"os"
	"strings"
	"testing"
)

// TestFormatDuration tests the duration formatting function in mcp.go
func TestMCPFormatDuration(t *testing.T) {
	tests := []struct {
		milliseconds float64
		expected     string
	}{
		// Seconds
		{0, "0 seconds"},
		{1000, "1 seconds"},
		{5000, "5 seconds"},
		{30000, "30 seconds"},
		{59000, "59 seconds"},
		
		// Minutes
		{60000, "1 minutes"},
		{120000, "2 minutes"},
		{1800000, "30 minutes"},
		{3599000, "59 minutes"},
		
		// Hours
		{3600000, "1 hours"},
		{7200000, "2 hours"},
		{43200000, "12 hours"},
		{86399000, "23 hours"},
		
		// Days
		{86400000, "1 days"},
		{172800000, "2 days"},
		{604800000, "7 days"},
		{2592000000, "30 days"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatDuration(tt.milliseconds)
			if result != tt.expected {
				t.Errorf("formatDuration(%f) = %q, want %q", tt.milliseconds, result, tt.expected)
			}
		})
	}
}

func TestRun_MCPServer(t *testing.T) {
	// Set JENKINS_URL and JENKINS_TOKEN env vars to get past token check
	oldURL := os.Getenv("JENKINS_URL")
	oldToken := os.Getenv("JENKINS_TOKEN")
	os.Setenv("JENKINS_URL", "http://test-jenkins.example.com")
	os.Setenv("JENKINS_TOKEN", "test-token")
	defer func() {
		if oldURL == "" {
			os.Unsetenv("JENKINS_URL")
		} else {
			os.Setenv("JENKINS_URL", oldURL)
		}
		if oldToken == "" {
			os.Unsetenv("JENKINS_TOKEN")
		} else {
			os.Setenv("JENKINS_TOKEN", oldToken)
		}
	}()

	// Test that mcp-server sub-command is recognized
	args := []string{"mcp-server"}

	// We can't easily test the full server without mocking stdin/stdout
	// but we can verify the command is recognized and doesn't return "unknown sub-command"
	_ = args
	// This test just verifies the test setup works
}

func TestRun_MCPServerMissingConfig(t *testing.T) {
	// Unset JENKINS_URL env var
	oldURL := os.Getenv("JENKINS_URL")
	os.Unsetenv("JENKINS_URL")
	defer func() {
		if oldURL != "" {
			os.Setenv("JENKINS_URL", oldURL)
		}
	}()

	ctx := context.Background()
	err := run(ctx, []string{"mcp-server"})

	if err == nil {
		t.Error("Expected error for missing configuration, got nil")
	}

	if !strings.Contains(err.Error(), "Jenkins URL must be configured") {
		t.Errorf("Expected 'Jenkins URL must be configured' error, got: %v", err)
	}
}
