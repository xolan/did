package timeutil

import (
	"testing"
	"time"
)

func TestParseDate_ISOFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
	}{
		{
			name:     "standard date",
			input:    "2024-01-15",
			expected: makeTime(2024, time.January, 15, 0, 0, 0),
		},
		{
			name:     "first day of year",
			input:    "2024-01-01",
			expected: makeTime(2024, time.January, 1, 0, 0, 0),
		},
		{
			name:     "last day of year",
			input:    "2024-12-31",
			expected: makeTime(2024, time.December, 31, 0, 0, 0),
		},
		{
			name:     "leap year feb 29",
			input:    "2024-02-29",
			expected: makeTime(2024, time.February, 29, 0, 0, 0),
		},
		{
			name:     "month boundary",
			input:    "2024-03-31",
			expected: makeTime(2024, time.March, 31, 0, 0, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDate(tt.input)
			if err != nil {
				t.Fatalf("ParseDate(%q) unexpected error: %v", tt.input, err)
			}
			if !result.Equal(tt.expected) {
				t.Errorf("ParseDate(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
			// Verify it's midnight
			if result.Hour() != 0 || result.Minute() != 0 || result.Second() != 0 || result.Nanosecond() != 0 {
				t.Errorf("ParseDate(%q) not midnight: got %02d:%02d:%02d.%09d",
					tt.input, result.Hour(), result.Minute(), result.Second(), result.Nanosecond())
			}
		})
	}
}

func TestParseDate_EuropeanFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
	}{
		{
			name:     "standard date",
			input:    "15/01/2024",
			expected: makeTime(2024, time.January, 15, 0, 0, 0),
		},
		{
			name:     "first day of year",
			input:    "01/01/2024",
			expected: makeTime(2024, time.January, 1, 0, 0, 0),
		},
		{
			name:     "last day of year",
			input:    "31/12/2024",
			expected: makeTime(2024, time.December, 31, 0, 0, 0),
		},
		{
			name:     "leap year feb 29",
			input:    "29/02/2024",
			expected: makeTime(2024, time.February, 29, 0, 0, 0),
		},
		{
			name:     "month boundary",
			input:    "31/03/2024",
			expected: makeTime(2024, time.March, 31, 0, 0, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDate(tt.input)
			if err != nil {
				t.Fatalf("ParseDate(%q) unexpected error: %v", tt.input, err)
			}
			if !result.Equal(tt.expected) {
				t.Errorf("ParseDate(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
			// Verify it's midnight
			if result.Hour() != 0 || result.Minute() != 0 || result.Second() != 0 || result.Nanosecond() != 0 {
				t.Errorf("ParseDate(%q) not midnight: got %02d:%02d:%02d.%09d",
					tt.input, result.Hour(), result.Minute(), result.Second(), result.Nanosecond())
			}
		})
	}
}

func TestParseDate_AmbiguousDates(t *testing.T) {
	// Test that ambiguous dates like 05/06/2024 prefer ISO format
	// In ISO format, this would be May 6, 2024
	// In European format, this would be June 5, 2024
	// We prefer ISO, so we expect May 6
	tests := []struct {
		name        string
		input       string
		expectedDay int
		description string
	}{
		{
			name:        "05-06-2024 is May 6 (ISO)",
			input:       "2024-05-06",
			expectedDay: 6,
			description: "ISO format should interpret as year-month-day",
		},
		{
			name:        "06/05/2024 is May 6 (European)",
			input:       "06/05/2024",
			expectedDay: 6,
			description: "European format should interpret as day/month/year",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDate(tt.input)
			if err != nil {
				t.Fatalf("ParseDate(%q) unexpected error: %v", tt.input, err)
			}
			if result.Day() != tt.expectedDay {
				t.Errorf("ParseDate(%q) day = %d, expected %d (%s)",
					tt.input, result.Day(), tt.expectedDay, tt.description)
			}
			if result.Month() != time.May {
				t.Errorf("ParseDate(%q) month = %v, expected May", tt.input, result.Month())
			}
		})
	}
}

func TestParseDate_InvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "invalid format",
			input: "invalid",
		},
		{
			name:  "partial date",
			input: "2024-01",
		},
		{
			name:  "wrong separator",
			input: "2024.01.15",
		},
		{
			name:  "US format not supported",
			input: "01/15/2024",
		},
		{
			name:  "plain text",
			input: "January 15, 2024",
		},
		{
			name:  "invalid day",
			input: "2024-02-30",
		},
		{
			name:  "invalid month",
			input: "2024-13-01",
		},
		{
			name:  "non-leap year feb 29",
			input: "2023-02-29",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDate(tt.input)
			if err == nil {
				t.Errorf("ParseDate(%q) expected error, got result: %v", tt.input, result)
			}
			// Verify zero time is returned on error
			if !result.IsZero() {
				t.Errorf("ParseDate(%q) expected zero time on error, got: %v", tt.input, result)
			}
		})
	}
}

func TestParseDate_BoundaryDates(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
	}{
		{
			name:     "year boundary - last day of year",
			input:    "2023-12-31",
			expected: makeTime(2023, time.December, 31, 0, 0, 0),
		},
		{
			name:     "year boundary - first day of year",
			input:    "2024-01-01",
			expected: makeTime(2024, time.January, 1, 0, 0, 0),
		},
		{
			name:     "month boundary - end of January",
			input:    "2024-01-31",
			expected: makeTime(2024, time.January, 31, 0, 0, 0),
		},
		{
			name:     "month boundary - start of February",
			input:    "2024-02-01",
			expected: makeTime(2024, time.February, 1, 0, 0, 0),
		},
		{
			name:     "short month - end of February",
			input:    "2024-02-29",
			expected: makeTime(2024, time.February, 29, 0, 0, 0),
		},
		{
			name:     "30-day month - end of April",
			input:    "2024-04-30",
			expected: makeTime(2024, time.April, 30, 0, 0, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDate(tt.input)
			if err != nil {
				t.Fatalf("ParseDate(%q) unexpected error: %v", tt.input, err)
			}
			if !result.Equal(tt.expected) {
				t.Errorf("ParseDate(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseRelativeDays_ValidInput(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedDays int
	}{
		{
			name:         "last 1 day",
			input:        "last 1 day",
			expectedDays: 1,
		},
		{
			name:         "last 7 days",
			input:        "last 7 days",
			expectedDays: 7,
		},
		{
			name:         "last 30 days",
			input:        "last 30 days",
			expectedDays: 30,
		},
		{
			name:         "last 90 days",
			input:        "last 90 days",
			expectedDays: 90,
		},
		{
			name:         "last 365 days",
			input:        "last 365 days",
			expectedDays: 365,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := ParseRelativeDays(tt.input)
			if err != nil {
				t.Fatalf("ParseRelativeDays(%q) unexpected error: %v", tt.input, err)
			}

			// Verify start is midnight
			if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
				t.Errorf("ParseRelativeDays(%q) start not midnight: got %02d:%02d:%02d",
					tt.input, start.Hour(), start.Minute(), start.Second())
			}

			// Verify end is end of day
			if end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 {
				t.Errorf("ParseRelativeDays(%q) end not end of day: got %02d:%02d:%02d",
					tt.input, end.Hour(), end.Minute(), end.Second())
			}

			// Verify the range covers the expected number of days
			// For ranges spanning DST transitions, we check the date span rather than exact duration
			startDate := StartOfDay(start)
			endDate := StartOfDay(end)
			daysDiff := int(endDate.Sub(startDate).Hours()/24) + 1 // +1 because range is inclusive

			if daysDiff != tt.expectedDays {
				t.Errorf("ParseRelativeDays(%q) covers %d days, expected %d (start=%v, end=%v)",
					tt.input, daysDiff, tt.expectedDays, start, end)
			}

			// Verify start is N-1 days before end (in terms of calendar days)
			expectedStart := StartOfDay(EndOfDay(time.Now()).AddDate(0, 0, -(tt.expectedDays - 1)))
			if !start.Equal(expectedStart) {
				t.Errorf("ParseRelativeDays(%q) start = %v, expected %v",
					tt.input, start, expectedStart)
			}

			// Verify end is today's end
			expectedEnd := EndOfDay(time.Now())
			if !end.Equal(expectedEnd) {
				t.Errorf("ParseRelativeDays(%q) end = %v, expected %v",
					tt.input, end, expectedEnd)
			}
		})
	}
}

func TestParseRelativeDays_SingularForm(t *testing.T) {
	// Test that singular "day" is also supported
	start, end, err := ParseRelativeDays("last 1 day")
	if err != nil {
		t.Fatalf("ParseRelativeDays('last 1 day') unexpected error: %v", err)
	}

	// Should return today only
	expectedStart := StartOfDay(time.Now())
	expectedEnd := EndOfDay(time.Now())

	if !start.Equal(expectedStart) {
		t.Errorf("ParseRelativeDays('last 1 day') start = %v, expected %v", start, expectedStart)
	}
	if !end.Equal(expectedEnd) {
		t.Errorf("ParseRelativeDays('last 1 day') end = %v, expected %v", end, expectedEnd)
	}
}

func TestParseRelativeDays_InvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "missing number",
			input: "last days",
		},
		{
			name:  "missing 'last'",
			input: "7 days",
		},
		{
			name:  "missing 'days'",
			input: "last 7",
		},
		{
			name:  "negative number",
			input: "last -7 days",
		},
		{
			name:  "zero days",
			input: "last 0 days",
		},
		{
			name:  "invalid format",
			input: "previous 7 days",
		},
		{
			name:  "non-numeric",
			input: "last seven days",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := ParseRelativeDays(tt.input)
			if err == nil {
				t.Errorf("ParseRelativeDays(%q) expected error, got start=%v, end=%v", tt.input, start, end)
			}
			// Verify zero times are returned on error
			if !start.IsZero() || !end.IsZero() {
				t.Errorf("ParseRelativeDays(%q) expected zero times on error, got start=%v, end=%v",
					tt.input, start, end)
			}
		})
	}
}

func TestParseRelativeDays_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedDays int
		description  string
	}{
		{
			name:         "last 1 day is today",
			input:        "last 1 day",
			expectedDays: 1,
			description:  "Should return only today",
		},
		{
			name:         "last 2 days",
			input:        "last 2 days",
			expectedDays: 2,
			description:  "Should return yesterday and today",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := ParseRelativeDays(tt.input)
			if err != nil {
				t.Fatalf("ParseRelativeDays(%q) unexpected error: %v", tt.input, err)
			}

			// Calculate actual number of days in range
			duration := end.Sub(start)
			actualDays := int(duration.Hours()/24) + 1 // +1 because range is inclusive

			if actualDays != tt.expectedDays {
				t.Errorf("ParseRelativeDays(%q) covers %d days, expected %d (%s)",
					tt.input, actualDays, tt.expectedDays, tt.description)
			}
		})
	}
}

func TestParseDate_PartialDateErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name:          "year only",
			input:         "2024",
			expectedError: "incomplete date '2024': missing month and day",
		},
		{
			name:          "ISO partial - missing day",
			input:         "2024-01",
			expectedError: "incomplete date '2024-01': missing day",
		},
		{
			name:          "ISO partial - missing year",
			input:         "01-15",
			expectedError: "incomplete date '01-15': missing year",
		},
		{
			name:          "European partial - missing year",
			input:         "15/01",
			expectedError: "incomplete date '15/01': missing year",
		},
		{
			name:          "empty string",
			input:         "",
			expectedError: "date cannot be empty",
		},
		{
			name:          "invalid format",
			input:         "invalid",
			expectedError: "invalid date format 'invalid'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseDate(tt.input)
			if err == nil {
				t.Fatalf("ParseDate(%q) expected error, got nil", tt.input)
			}
			if !containsSubstring(err.Error(), tt.expectedError) {
				t.Errorf("ParseDate(%q) error = %q, expected to contain %q",
					tt.input, err.Error(), tt.expectedError)
			}
		})
	}
}

func TestParseRelativeDays_ErrorMessages(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name:          "empty string",
			input:         "",
			expectedError: "relative date cannot be empty",
		},
		{
			name:          "invalid format",
			input:         "last week",
			expectedError: "invalid format 'last week'",
		},
		{
			name:          "missing last",
			input:         "7 days",
			expectedError: "invalid format '7 days'",
		},
		{
			name:          "missing days",
			input:         "last 7",
			expectedError: "invalid format 'last 7'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := ParseRelativeDays(tt.input)
			if err == nil {
				t.Fatalf("ParseRelativeDays(%q) expected error, got nil", tt.input)
			}
			if !containsSubstring(err.Error(), tt.expectedError) {
				t.Errorf("ParseRelativeDays(%q) error = %q, expected to contain %q",
					tt.input, err.Error(), tt.expectedError)
			}
		})
	}
}

// containsSubstring checks if s contains substr
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestParseDate_TooManyParts(t *testing.T) {
	// Test the tooManyPartsRe case in buildDateParseError
	tests := []struct {
		name           string
		input          string
		expectedSubstr string
	}{
		{
			name:           "too many hyphens",
			input:          "2024-01-15-01",
			expectedSubstr: "too many date parts",
		},
		{
			name:           "too many slashes",
			input:          "15/01/2024/12",
			expectedSubstr: "too many date parts",
		},
		{
			name:           "mixed separators with extra",
			input:          "2024-01-15-",
			expectedSubstr: "too many date parts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseDate(tt.input)
			if err == nil {
				t.Fatalf("ParseDate(%q) expected error, got nil", tt.input)
			}
			if !containsSubstring(err.Error(), tt.expectedSubstr) {
				t.Errorf("ParseDate(%q) error = %q, expected to contain %q",
					tt.input, err.Error(), tt.expectedSubstr)
			}
		})
	}
}
