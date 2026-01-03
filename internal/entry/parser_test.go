package entry

import (
	"strings"
	"testing"
)

func TestParseDuration_Hours(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"1 hour", "1h", 60},
		{"2 hours", "2h", 120},
		{"10 hours", "10h", 600},
		{"24 hours (max)", "24h", 1440},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDuration(tt.input)
			if err != nil {
				t.Errorf("ParseDuration(%q) returned unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("ParseDuration(%q) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseDuration_Minutes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"1 minute", "1m", 1},
		{"30 minutes", "30m", 30},
		{"45 minutes", "45m", 45},
		{"60 minutes", "60m", 60},
		{"90 minutes", "90m", 90},
		{"1440 minutes (24h max)", "1440m", 1440},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDuration(tt.input)
			if err != nil {
				t.Errorf("ParseDuration(%q) returned unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("ParseDuration(%q) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseDuration_CombinedFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"1h30m", "1h30m", 90},
		{"2h15m", "2h15m", 135},
		{"0h30m", "0h30m", 30},
		{"1h0m", "1h0m", 60},
		{"10h45m", "10h45m", 645},
		{"23h59m", "23h59m", 1439},
		{"24h0m", "24h0m", 1440},
		{"0h1m", "0h1m", 1},
		{"5h5m", "5h5m", 305},
		{"12h30m", "12h30m", 750},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDuration(tt.input)
			if err != nil {
				t.Errorf("ParseDuration(%q) returned unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("ParseDuration(%q) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseDuration_Invalid(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		errorSubstring string
	}{
		{"no unit", "2", "invalid time format"},
		{"invalid unit", "2x", "invalid time format"},
		{"text only", "invalid", "invalid time format"},
		{"empty string", "", "invalid time format"},
		{"just hour unit", "h", "invalid time format"},
		{"just minute unit", "m", "invalid time format"},
		{"negative hours", "-2h", "invalid time format"},
		{"decimal hours", "1.5h", "invalid time format"},
		{"space in input", "2 h", "invalid time format"},
		{"hours then text", "2hours", "invalid time format"},
		{"minutes then text", "30minutes", "invalid time format"},
		{"mixed case", "2H", "invalid time format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDuration(tt.input)
			if err == nil {
				t.Errorf("ParseDuration(%q) = %d, expected error containing %q", tt.input, result, tt.errorSubstring)
			} else if !strings.Contains(err.Error(), tt.errorSubstring) {
				t.Errorf("ParseDuration(%q) error = %q, expected to contain %q", tt.input, err.Error(), tt.errorSubstring)
			}
		})
	}
}

func TestParseDuration_Zero(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		errorSubstring string
	}{
		{"zero hours", "0h", "duration cannot be zero"},
		{"zero minutes", "0m", "duration cannot be zero"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDuration(tt.input)
			if err == nil {
				t.Errorf("ParseDuration(%q) = %d, expected error containing %q", tt.input, result, tt.errorSubstring)
			} else if !strings.Contains(err.Error(), tt.errorSubstring) {
				t.Errorf("ParseDuration(%q) error = %q, expected to contain %q", tt.input, err.Error(), tt.errorSubstring)
			}
		})
	}
}

func TestParseDuration_ExceedsMax(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		errorSubstring string
	}{
		{"25 hours", "25h", "exceeds maximum"},
		{"48 hours", "48h", "exceeds maximum"},
		{"1441 minutes", "1441m", "exceeds maximum"},
		{"2000 minutes", "2000m", "exceeds maximum"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDuration(tt.input)
			if err == nil {
				t.Errorf("ParseDuration(%q) = %d, expected error containing %q", tt.input, result, tt.errorSubstring)
			} else if !strings.Contains(err.Error(), tt.errorSubstring) {
				t.Errorf("ParseDuration(%q) error = %q, expected to contain %q", tt.input, err.Error(), tt.errorSubstring)
			}
		})
	}
}

func TestParseDuration_InvalidCombinedFormat(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		errorSubstring string
	}{
		// Exceeds max duration
		{"25h30m exceeds max", "25h30m", "exceeds maximum"},
		{"24h1m exceeds max", "24h1m", "exceeds maximum"},
		{"30h0m exceeds max", "30h0m", "exceeds maximum"},
		{"50h45m exceeds max", "50h45m", "exceeds maximum"},

		// Zero duration
		{"0h0m zero duration", "0h0m", "duration cannot be zero"},

		// Malformed patterns
		{"missing minute unit", "1h30", "invalid time format"},
		{"missing hour unit", "1 30m", "invalid time format"},
		{"wrong order", "30m1h", "invalid time format"},
		{"space between", "1h 30m", "invalid time format"},
		{"negative hours", "-1h30m", "invalid time format"},
		{"negative minutes", "1h-30m", "invalid time format"},
		{"decimal hours", "1.5h30m", "invalid time format"},
		{"decimal minutes", "1h30.5m", "invalid time format"},
		{"double hour unit", "1hh", "invalid time format"},
		{"double minute unit", "1mm", "invalid time format"},
		{"uppercase H", "1H30m", "invalid time format"},
		{"uppercase M", "1h30M", "invalid time format"},
		{"uppercase both", "1H30M", "invalid time format"},
		{"extra text after", "1h30minutes", "invalid time format"},
		{"extra text before", "time1h30m", "invalid time format"},
		{"only h and m", "hm", "invalid time format"},
		{"reversed units", "1m30h", "invalid time format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDuration(tt.input)
			if err == nil {
				t.Errorf("ParseDuration(%q) = %d, expected error containing %q", tt.input, result, tt.errorSubstring)
			} else if !strings.Contains(err.Error(), tt.errorSubstring) {
				t.Errorf("ParseDuration(%q) error = %q, expected to contain %q", tt.input, err.Error(), tt.errorSubstring)
			}
		})
	}
}

func TestMaxDurationMinutes(t *testing.T) {
	// Verify the constant is correctly set to 24 hours
	expected := 24 * 60
	if MaxDurationMinutes != expected {
		t.Errorf("MaxDurationMinutes = %d, expected %d (24 hours)", MaxDurationMinutes, expected)
	}
}
