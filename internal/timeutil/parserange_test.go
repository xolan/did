package timeutil

import (
	"strings"
	"testing"
	"time"
)

func TestParseDateRangeFlags_LastDays(t *testing.T) {
	tests := []struct {
		name     string
		lastDays int
		wantErr  bool
	}{
		{"last 1 day", 1, false},
		{"last 7 days", 7, false},
		{"last 30 days", 30, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := ParseDateRangeFlags("", "", tt.lastDays)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDateRangeFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			now := time.Now()
			expectedEnd := EndOfDay(now)
			expectedStart := StartOfDay(now.AddDate(0, 0, -(tt.lastDays - 1)))

			if !start.Equal(expectedStart) {
				t.Errorf("start = %v, want %v", start, expectedStart)
			}
			if !end.Equal(expectedEnd) {
				t.Errorf("end = %v, want %v", end, expectedEnd)
			}
		})
	}
}

func TestParseDateRangeFlags_FromTo(t *testing.T) {
	tests := []struct {
		name    string
		from    string
		to      string
		wantErr bool
	}{
		{"from only", "2024-01-01", "", false},
		{"to only", "", "2024-01-31", false},
		{"from and to", "2024-01-01", "2024-01-31", false},
		{"invalid from", "invalid", "", true},
		{"invalid to", "", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := ParseDateRangeFlags(tt.from, tt.to, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDateRangeFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if tt.from != "" && start.IsZero() {
				t.Error("expected non-zero start time")
			}
			if !end.IsZero() && end.Hour() != 23 {
				t.Errorf("end should be end of day, got hour %d", end.Hour())
			}
		})
	}
}

func TestParseDateRangeFlags_Conflict(t *testing.T) {
	_, _, err := ParseDateRangeFlags("2024-01-01", "", 7)
	if err == nil {
		t.Error("expected error when using --last with --from")
	}
	if !strings.Contains(err.Error(), "cannot use --last with --from or --to") {
		t.Errorf("unexpected error message: %v", err)
	}

	_, _, err = ParseDateRangeFlags("", "2024-01-31", 7)
	if err == nil {
		t.Error("expected error when using --last with --to")
	}
}

func TestParseDateRangeFlags_FromAfterTo(t *testing.T) {
	_, _, err := ParseDateRangeFlags("2024-12-31", "2024-01-01", 0)
	if err == nil {
		t.Error("expected error when from is after to")
	}
	if !strings.Contains(err.Error(), "is after") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestParseDateRangeFlags_NoFlags(t *testing.T) {
	start, end, err := ParseDateRangeFlags("", "", 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !start.IsZero() {
		t.Error("expected zero start time when no flags provided")
	}
	now := time.Now()
	expectedEnd := EndOfDay(now)
	if !end.Equal(expectedEnd) {
		t.Errorf("end = %v, want %v", end, expectedEnd)
	}
}
