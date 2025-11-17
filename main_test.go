package main

import (
	"context"
	"strings"
	"testing"
)

// TestGetStatusFromColor tests the color to status conversion
func TestGetStatusFromColor(t *testing.T) {
	tests := []struct {
		color    string
		expected string
	}{
		{"", ""}, // Empty color for non-buildable jobs
		{"blue", "SUCCESS"},
		{"blue_anime", "SUCCESS"},
		{"red", "FAILURE"},
		{"red_anime", "FAILURE"},
		{"yellow", "UNSTABLE"},
		{"yellow_anime", "UNSTABLE"},
		{"grey", "PENDING"},
		{"grey_anime", "PENDING"},
		{"aborted", "ABORTED"},
		{"aborted_anime", "ABORTED"},
		{"notbuilt", "NOT_BUILT"},
		{"disabled", "DISABLED"},
		{"disabled_anime", "DISABLED"},
	}

	for _, tt := range tests {
		t.Run(tt.color, func(t *testing.T) {
			result := getStatusFromColor(tt.color)
			if result != tt.expected {
				t.Errorf("getStatusFromColor(%q) = %q, want %q", tt.color, result, tt.expected)
			}
		})
	}
}

// TestRun_BuildJobRejected verifies that the build-job command is properly rejected
func TestRun_BuildJobRejected(t *testing.T) {
	ctx := context.Background()
	err := run(ctx, []string{"build-job", "test-job"})

	if err == nil {
		t.Fatal("Expected error for build-job command, got nil")
	}

	if !strings.Contains(err.Error(), "unknown sub-command: build-job") {
		t.Errorf("Expected 'unknown sub-command: build-job' error, got: %v", err)
	}
}
