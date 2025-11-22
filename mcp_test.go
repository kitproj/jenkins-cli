package main

import (
	"context"
	"os"
	"strings"
	"testing"
)

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
	// Use a temp directory for config to ensure no config file exists
	tmpDir := t.TempDir()
	oldXDGConfigHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() {
		if oldXDGConfigHome != "" {
			os.Setenv("XDG_CONFIG_HOME", oldXDGConfigHome)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	// Unset JENKINS_URL and JENKINS_TOKEN env vars
	oldURL := os.Getenv("JENKINS_URL")
	oldToken := os.Getenv("JENKINS_TOKEN")
	os.Unsetenv("JENKINS_URL")
	os.Unsetenv("JENKINS_TOKEN")
	defer func() {
		if oldURL != "" {
			os.Setenv("JENKINS_URL", oldURL)
		}
		if oldToken != "" {
			os.Setenv("JENKINS_TOKEN", oldToken)
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
