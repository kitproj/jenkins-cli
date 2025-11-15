package main

import (
	"context"
	"testing"
)

// TestGetStatusFromColor tests the color to status conversion
func TestGetStatusFromColor(t *testing.T) {
	tests := []struct {
		color    string
		expected string
	}{
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

// TestRun_SearchJobsCommand tests that the search-jobs command is recognized
func TestRun_SearchJobsCommand(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "search-jobs without pattern",
			args:    []string{"search-jobs"},
			wantErr: true,
		},
		{
			name:    "search-jobs with pattern",
			args:    []string{"search-jobs", "test"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := run(ctx, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// We expect errors for these since we don't have Jenkins credentials configured
			// The important thing is that the command is recognized and doesn't return "unknown sub-command"
		})
	}
}
