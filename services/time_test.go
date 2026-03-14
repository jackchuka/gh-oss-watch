package services

import (
	"testing"
	"time"
)

func TestHumanizeAge(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"zero time", 0, "—"},
		{"just now", 30 * time.Second, "just now"},
		{"1 minute", 90 * time.Second, "1 minute"},
		{"5 minutes", 5 * time.Minute, "5 minutes"},
		{"1 hour", 90 * time.Minute, "1 hour"},
		{"3 hours", 3 * time.Hour, "3 hours"},
		{"1 day", 36 * time.Hour, "1 day"},
		{"5 days", 5 * 24 * time.Hour, "5 days"},
		{"1 week", 10 * 24 * time.Hour, "1 week"},
		{"3 weeks", 21 * 24 * time.Hour, "3 weeks"},
		{"1 month", 35 * 24 * time.Hour, "1 month"},
		{"6 months", 180 * 24 * time.Hour, "6 months"},
		{"1 year", 400 * 24 * time.Hour, "1 year"},
		{"3 years", 3 * 365 * 24 * time.Hour, "3 years"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input time.Time
			if tt.duration == 0 {
				input = time.Time{} // zero value
			} else {
				input = time.Now().Add(-tt.duration)
			}
			got := HumanizeAge(input)
			if got != tt.expected {
				t.Errorf("HumanizeAge(%v ago) = %q, want %q", tt.duration, got, tt.expected)
			}
		})
	}
}
