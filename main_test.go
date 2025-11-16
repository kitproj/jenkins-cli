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

// TestParseJobPath tests the job path parsing function
func TestParseJobPath(t *testing.T) {
	tests := []struct {
		name           string
		jobPath        string
		expectedJob    string
		expectedParents []string
	}{
		{
			name:           "simple job name",
			jobPath:        "simple-job",
			expectedJob:    "simple-job",
			expectedParents: nil,
		},
		{
			name:           "nested job with one parent",
			jobPath:        "folder1/job/my-job",
			expectedJob:    "my-job",
			expectedParents: []string{"folder1"},
		},
		{
			name:           "nested job with multiple parents",
			jobPath:        "cloud-workspaces/job/cws-api/job/cws-api-a/job/master",
			expectedJob:    "master",
			expectedParents: []string{"cloud-workspaces", "cws-api", "cws-api-a"},
		},
		{
			name:           "two level nesting",
			jobPath:        "folder1/job/folder2/job/my-job",
			expectedJob:    "my-job",
			expectedParents: []string{"folder1", "folder2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job, parents := parseJobPath(tt.jobPath)
			if job != tt.expectedJob {
				t.Errorf("parseJobPath(%q) job = %q, want %q", tt.jobPath, job, tt.expectedJob)
			}
			if len(parents) != len(tt.expectedParents) {
				t.Errorf("parseJobPath(%q) parents length = %d, want %d", tt.jobPath, len(parents), len(tt.expectedParents))
				return
			}
			for i, parent := range parents {
				if parent != tt.expectedParents[i] {
					t.Errorf("parseJobPath(%q) parents[%d] = %q, want %q", tt.jobPath, i, parent, tt.expectedParents[i])
				}
			}
		})
	}
}
