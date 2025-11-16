package main

import (
	"testing"
)

// TestGetStatusFromColor tests the color to status conversion
func TestGetStatusFromColor(t *testing.T) {
	tests := []struct {
		color    string
		expected string
	}{
		{"", "N/A"}, // Empty color for non-buildable jobs
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
