package main

import (
	"testing"
)

// TestFormatDuration tests the duration formatting function
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		milliseconds float64
		expected     string
	}{
		// Seconds
		{0, "0 seconds"},
		{1000, "1 second"},
		{5000, "5 seconds"},
		{30000, "30 seconds"},
		{59000, "59 seconds"},
		
		// Minutes
		{60000, "1 minute"},
		{120000, "2 minutes"},
		{1800000, "30 minutes"},
		{3599000, "59 minutes"},
		
		// Hours
		{3600000, "1 hour"},
		{7200000, "2 hours"},
		{43200000, "12 hours"},
		{86399000, "23 hours"},
		
		// Days
		{86400000, "1 day"},
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
