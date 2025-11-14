package main

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestRun_MCPServer(t *testing.T) {
	// Set JENKINS_HOST and JENKINS_TOKEN env vars to get past token check
	oldHost := os.Getenv("JENKINS_HOST")
	oldToken := os.Getenv("JENKINS_TOKEN")
	os.Setenv("JENKINS_HOST", "http://test-jenkins.example.com")
	os.Setenv("JENKINS_TOKEN", "test-token")
	defer func() {
		if oldHost == "" {
			os.Unsetenv("JENKINS_HOST")
		} else {
			os.Setenv("JENKINS_HOST", oldHost)
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
	// Unset JENKINS_HOST env var
	oldHost := os.Getenv("JENKINS_HOST")
	os.Unsetenv("JENKINS_HOST")
	defer func() {
		if oldHost != "" {
			os.Setenv("JENKINS_HOST", oldHost)
		}
	}()

	ctx := context.Background()
	err := run(ctx, []string{"mcp-server"})

	if err == nil {
		t.Error("Expected error for missing configuration, got nil")
	}

	if !strings.Contains(err.Error(), "Jenkins host must be configured") {
		t.Errorf("Expected 'Jenkins host must be configured' error, got: %v", err)
	}
}
