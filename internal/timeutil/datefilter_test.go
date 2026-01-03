package timeutil

import (
	"testing"
	"time"
)

// Helper function to create test times with specific dates
func makeTime(year int, month time.Month, day, hour, min, sec int) time.Time {
	return time.Date(year, month, day, hour, min, sec, 0, time.Local)
}

func TestStartOfDay(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "midnight stays midnight",
			input:    makeTime(2024, time.January, 15, 0, 0, 0),
			expected: makeTime(2024, time.January, 15, 0, 0, 0),
		},
		{
			name:     "noon becomes midnight",
			input:    makeTime(2024, time.January, 15, 12, 0, 0),
			expected: makeTime(2024, time.January, 15, 0, 0, 0),
		},
		{
			name:     "end of day becomes midnight",
			input:    makeTime(2024, time.January, 15, 23, 59, 59),
			expected: makeTime(2024, time.January, 15, 0, 0, 0),
		},
		{
			name:     "random time becomes midnight",
			input:    makeTime(2024, time.March, 20, 14, 35, 22),
			expected: makeTime(2024, time.March, 20, 0, 0, 0),
		},
		{
			name:     "new year's day",
			input:    makeTime(2024, time.January, 1, 10, 30, 0),
			expected: makeTime(2024, time.January, 1, 0, 0, 0),
		},
		{
			name:     "leap year feb 29",
			input:    makeTime(2024, time.February, 29, 18, 45, 30),
			expected: makeTime(2024, time.February, 29, 0, 0, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StartOfDay(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("StartOfDay(%v) = %v, expected %v", tt.input, result, tt.expected)
			}
			// Verify all time components are zero
			if result.Hour() != 0 || result.Minute() != 0 || result.Second() != 0 || result.Nanosecond() != 0 {
				t.Errorf("StartOfDay(%v) time components not all zero: got %02d:%02d:%02d.%09d",
					tt.input, result.Hour(), result.Minute(), result.Second(), result.Nanosecond())
			}
		})
	}
}

func TestEndOfDay(t *testing.T) {
	tests := []struct {
		name  string
		input time.Time
	}{
		{
			name:  "midnight",
			input: makeTime(2024, time.January, 15, 0, 0, 0),
		},
		{
			name:  "noon",
			input: makeTime(2024, time.January, 15, 12, 0, 0),
		},
		{
			name:  "end of day",
			input: makeTime(2024, time.January, 15, 23, 59, 59),
		},
		{
			name:  "leap year feb 29",
			input: makeTime(2024, time.February, 29, 18, 45, 30),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EndOfDay(tt.input)

			// Should be same day
			if result.Year() != tt.input.Year() || result.Month() != tt.input.Month() || result.Day() != tt.input.Day() {
				t.Errorf("EndOfDay(%v) changed the date: got %v", tt.input, result)
			}

			// Should be 23:59:59.999999999
			if result.Hour() != 23 || result.Minute() != 59 || result.Second() != 59 || result.Nanosecond() != 999999999 {
				t.Errorf("EndOfDay(%v) = %02d:%02d:%02d.%09d, expected 23:59:59.999999999",
					tt.input, result.Hour(), result.Minute(), result.Second(), result.Nanosecond())
			}
		})
	}
}

func TestStartOfWeek(t *testing.T) {
	tests := []struct {
		name           string
		input          time.Time
		expectedMonday time.Time
	}{
		{
			name:           "Monday stays Monday",
			input:          makeTime(2024, time.January, 15, 10, 30, 0), // Monday
			expectedMonday: makeTime(2024, time.January, 15, 0, 0, 0),
		},
		{
			name:           "Tuesday goes to Monday",
			input:          makeTime(2024, time.January, 16, 14, 0, 0), // Tuesday
			expectedMonday: makeTime(2024, time.January, 15, 0, 0, 0),
		},
		{
			name:           "Wednesday goes to Monday",
			input:          makeTime(2024, time.January, 17, 9, 15, 0), // Wednesday
			expectedMonday: makeTime(2024, time.January, 15, 0, 0, 0),
		},
		{
			name:           "Thursday goes to Monday",
			input:          makeTime(2024, time.January, 18, 16, 45, 0), // Thursday
			expectedMonday: makeTime(2024, time.January, 15, 0, 0, 0),
		},
		{
			name:           "Friday goes to Monday",
			input:          makeTime(2024, time.January, 19, 11, 0, 0), // Friday
			expectedMonday: makeTime(2024, time.January, 15, 0, 0, 0),
		},
		{
			name:           "Saturday goes to Monday",
			input:          makeTime(2024, time.January, 20, 8, 30, 0), // Saturday
			expectedMonday: makeTime(2024, time.January, 15, 0, 0, 0),
		},
		{
			name:           "Sunday goes to previous Monday",
			input:          makeTime(2024, time.January, 21, 20, 0, 0), // Sunday
			expectedMonday: makeTime(2024, time.January, 15, 0, 0, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StartOfWeek(tt.input)
			if !result.Equal(tt.expectedMonday) {
				t.Errorf("StartOfWeek(%v [%s]) = %v [%s], expected %v [Monday]",
					tt.input, tt.input.Weekday(), result, result.Weekday(), tt.expectedMonday)
			}

			// Verify it's a Monday
			if result.Weekday() != time.Monday {
				t.Errorf("StartOfWeek(%v) weekday = %s, expected Monday", tt.input, result.Weekday())
			}

			// Verify it's midnight
			if result.Hour() != 0 || result.Minute() != 0 || result.Second() != 0 || result.Nanosecond() != 0 {
				t.Errorf("StartOfWeek(%v) not midnight: got %02d:%02d:%02d.%09d",
					tt.input, result.Hour(), result.Minute(), result.Second(), result.Nanosecond())
			}
		})
	}
}

func TestStartOfWeek_Sunday(t *testing.T) {
	// Special test case for Sunday edge case (Go's Weekday() returns 0 for Sunday)
	tests := []struct {
		name           string
		sunday         time.Time
		expectedMonday time.Time
	}{
		{
			name:           "first Sunday of 2024",
			sunday:         makeTime(2024, time.January, 7, 12, 0, 0),
			expectedMonday: makeTime(2024, time.January, 1, 0, 0, 0),
		},
		{
			name:           "Sunday at midnight",
			sunday:         makeTime(2024, time.January, 21, 0, 0, 0),
			expectedMonday: makeTime(2024, time.January, 15, 0, 0, 0),
		},
		{
			name:           "Sunday at end of day",
			sunday:         makeTime(2024, time.January, 21, 23, 59, 59),
			expectedMonday: makeTime(2024, time.January, 15, 0, 0, 0),
		},
		{
			name:           "Sunday crossing month boundary",
			sunday:         makeTime(2024, time.March, 3, 15, 30, 0), // Sunday March 3
			expectedMonday: makeTime(2024, time.February, 26, 0, 0, 0),
		},
		{
			name:           "Sunday crossing year boundary",
			sunday:         makeTime(2024, time.January, 7, 10, 0, 0), // First Sunday of Jan 2024
			expectedMonday: makeTime(2024, time.January, 1, 0, 0, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify input is actually a Sunday
			if tt.sunday.Weekday() != time.Sunday {
				t.Fatalf("Test setup error: %v is %s, not Sunday", tt.sunday, tt.sunday.Weekday())
			}

			result := StartOfWeek(tt.sunday)
			if !result.Equal(tt.expectedMonday) {
				t.Errorf("StartOfWeek(%v) = %v, expected %v", tt.sunday, result, tt.expectedMonday)
			}

			if result.Weekday() != time.Monday {
				t.Errorf("StartOfWeek(%v) weekday = %s, expected Monday", tt.sunday, result.Weekday())
			}
		})
	}
}

func TestEndOfWeek(t *testing.T) {
	tests := []struct {
		name           string
		input          time.Time
		expectedSunday time.Time
	}{
		{
			name:           "Monday to Sunday",
			input:          makeTime(2024, time.January, 15, 10, 0, 0), // Monday
			expectedSunday: makeTime(2024, time.January, 21, 23, 59, 59),
		},
		{
			name:           "Wednesday to Sunday",
			input:          makeTime(2024, time.January, 17, 14, 30, 0), // Wednesday
			expectedSunday: makeTime(2024, time.January, 21, 23, 59, 59),
		},
		{
			name:           "Sunday stays same Sunday",
			input:          makeTime(2024, time.January, 21, 8, 0, 0), // Sunday
			expectedSunday: makeTime(2024, time.January, 21, 23, 59, 59),
		},
		{
			name:           "Saturday to Sunday",
			input:          makeTime(2024, time.January, 20, 16, 0, 0), // Saturday
			expectedSunday: makeTime(2024, time.January, 21, 23, 59, 59),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EndOfWeek(tt.input)

			// Check the date components
			if result.Year() != tt.expectedSunday.Year() || result.Month() != tt.expectedSunday.Month() || result.Day() != tt.expectedSunday.Day() {
				t.Errorf("EndOfWeek(%v) date = %v, expected %v", tt.input, result, tt.expectedSunday)
			}

			// Verify it's a Sunday
			if result.Weekday() != time.Sunday {
				t.Errorf("EndOfWeek(%v) weekday = %s, expected Sunday", tt.input, result.Weekday())
			}

			// Should be 23:59:59.999999999
			if result.Hour() != 23 || result.Minute() != 59 || result.Second() != 59 || result.Nanosecond() != 999999999 {
				t.Errorf("EndOfWeek(%v) = %02d:%02d:%02d.%09d, expected 23:59:59.999999999",
					tt.input, result.Hour(), result.Minute(), result.Second(), result.Nanosecond())
			}
		})
	}
}

func TestToday(t *testing.T) {
	start, end := Today()
	now := time.Now()

	// Start should be midnight today
	if start.Year() != now.Year() || start.Month() != now.Month() || start.Day() != now.Day() {
		t.Errorf("Today() start date mismatch: got %v, expected today %v", start, now)
	}
	if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
		t.Errorf("Today() start not midnight: got %02d:%02d:%02d", start.Hour(), start.Minute(), start.Second())
	}

	// End should be end of today
	if end.Year() != now.Year() || end.Month() != now.Month() || end.Day() != now.Day() {
		t.Errorf("Today() end date mismatch: got %v, expected today %v", end, now)
	}
	if end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 {
		t.Errorf("Today() end not end of day: got %02d:%02d:%02d", end.Hour(), end.Minute(), end.Second())
	}

	// Start should be before end
	if !start.Before(end) {
		t.Errorf("Today() start %v not before end %v", start, end)
	}
}

func TestYesterday(t *testing.T) {
	start, end := Yesterday()
	now := time.Now()
	expectedDay := now.AddDate(0, 0, -1)

	// Start should be midnight yesterday
	if start.Year() != expectedDay.Year() || start.Month() != expectedDay.Month() || start.Day() != expectedDay.Day() {
		t.Errorf("Yesterday() start date mismatch: got %v, expected %v", start, expectedDay)
	}
	if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
		t.Errorf("Yesterday() start not midnight: got %02d:%02d:%02d", start.Hour(), start.Minute(), start.Second())
	}

	// End should be end of yesterday
	if end.Year() != expectedDay.Year() || end.Month() != expectedDay.Month() || end.Day() != expectedDay.Day() {
		t.Errorf("Yesterday() end date mismatch: got %v, expected %v", end, expectedDay)
	}
	if end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 {
		t.Errorf("Yesterday() end not end of day: got %02d:%02d:%02d", end.Hour(), end.Minute(), end.Second())
	}

	// Start should be before end
	if !start.Before(end) {
		t.Errorf("Yesterday() start %v not before end %v", start, end)
	}

	// Yesterday should be exactly one day before today
	todayStart, _ := Today()
	if !start.AddDate(0, 0, 1).Equal(todayStart) {
		t.Errorf("Yesterday() start + 1 day (%v) != Today() start (%v)", start.AddDate(0, 0, 1), todayStart)
	}
}

func TestThisWeek(t *testing.T) {
	start, end := ThisWeek()

	// Start should be Monday
	if start.Weekday() != time.Monday {
		t.Errorf("ThisWeek() start weekday = %s, expected Monday", start.Weekday())
	}
	if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
		t.Errorf("ThisWeek() start not midnight: got %02d:%02d:%02d", start.Hour(), start.Minute(), start.Second())
	}

	// End should be Sunday
	if end.Weekday() != time.Sunday {
		t.Errorf("ThisWeek() end weekday = %s, expected Sunday", end.Weekday())
	}
	if end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 {
		t.Errorf("ThisWeek() end not end of day: got %02d:%02d:%02d", end.Hour(), end.Minute(), end.Second())
	}

	// Start should be before end
	if !start.Before(end) {
		t.Errorf("ThisWeek() start %v not before end %v", start, end)
	}

	// Duration should be approximately 7 days
	duration := end.Sub(start)
	expectedDuration := 7*24*time.Hour - time.Nanosecond
	if duration != expectedDuration {
		t.Errorf("ThisWeek() duration = %v, expected %v", duration, expectedDuration)
	}
}

func TestLastWeek(t *testing.T) {
	start, end := LastWeek()
	thisWeekStart, _ := ThisWeek()

	// Start should be Monday
	if start.Weekday() != time.Monday {
		t.Errorf("LastWeek() start weekday = %s, expected Monday", start.Weekday())
	}
	if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
		t.Errorf("LastWeek() start not midnight: got %02d:%02d:%02d", start.Hour(), start.Minute(), start.Second())
	}

	// End should be Sunday
	if end.Weekday() != time.Sunday {
		t.Errorf("LastWeek() end weekday = %s, expected Sunday", end.Weekday())
	}
	if end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 {
		t.Errorf("LastWeek() end not end of day: got %02d:%02d:%02d", end.Hour(), end.Minute(), end.Second())
	}

	// Start should be before end
	if !start.Before(end) {
		t.Errorf("LastWeek() start %v not before end %v", start, end)
	}

	// Last week start + 7 days should equal this week start
	if !start.AddDate(0, 0, 7).Equal(thisWeekStart) {
		t.Errorf("LastWeek() start + 7 days (%v) != ThisWeek() start (%v)", start.AddDate(0, 0, 7), thisWeekStart)
	}

	// Duration should be approximately 7 days
	duration := end.Sub(start)
	expectedDuration := 7*24*time.Hour - time.Nanosecond
	if duration != expectedDuration {
		t.Errorf("LastWeek() duration = %v, expected %v", duration, expectedDuration)
	}
}

func TestIsInRange(t *testing.T) {
	start := makeTime(2024, time.January, 15, 0, 0, 0)
	end := makeTime(2024, time.January, 15, 23, 59, 59)

	tests := []struct {
		name     string
		testTime time.Time
		expected bool
	}{
		{
			name:     "exactly at start",
			testTime: start,
			expected: true,
		},
		{
			name:     "exactly at end",
			testTime: end,
			expected: true,
		},
		{
			name:     "in middle of range",
			testTime: makeTime(2024, time.January, 15, 12, 0, 0),
			expected: true,
		},
		{
			name:     "one nanosecond after start",
			testTime: start.Add(time.Nanosecond),
			expected: true,
		},
		{
			name:     "one nanosecond before end",
			testTime: end.Add(-time.Nanosecond),
			expected: true,
		},
		{
			name:     "one nanosecond before start",
			testTime: start.Add(-time.Nanosecond),
			expected: false,
		},
		{
			name:     "one nanosecond after end",
			testTime: end.Add(time.Nanosecond),
			expected: false,
		},
		{
			name:     "day before",
			testTime: makeTime(2024, time.January, 14, 12, 0, 0),
			expected: false,
		},
		{
			name:     "day after",
			testTime: makeTime(2024, time.January, 16, 12, 0, 0),
			expected: false,
		},
		{
			name:     "year before",
			testTime: makeTime(2023, time.January, 15, 12, 0, 0),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsInRange(tt.testTime, start, end)
			if result != tt.expected {
				t.Errorf("IsInRange(%v, %v, %v) = %v, expected %v",
					tt.testTime, start, end, result, tt.expected)
			}
		})
	}
}

func TestIsInRange_WeekRange(t *testing.T) {
	// Test using a week range (Monday-Sunday)
	start := makeTime(2024, time.January, 15, 0, 0, 0)  // Monday
	end := makeTime(2024, time.January, 21, 23, 59, 59) // Sunday

	tests := []struct {
		name     string
		testTime time.Time
		expected bool
	}{
		{
			name:     "Monday morning",
			testTime: makeTime(2024, time.January, 15, 8, 0, 0),
			expected: true,
		},
		{
			name:     "Wednesday afternoon",
			testTime: makeTime(2024, time.January, 17, 14, 30, 0),
			expected: true,
		},
		{
			name:     "Sunday evening",
			testTime: makeTime(2024, time.January, 21, 20, 0, 0),
			expected: true,
		},
		{
			name:     "Previous Sunday",
			testTime: makeTime(2024, time.January, 14, 23, 59, 59),
			expected: false,
		},
		{
			name:     "Next Monday",
			testTime: makeTime(2024, time.January, 22, 0, 0, 0),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsInRange(tt.testTime, start, end)
			if result != tt.expected {
				t.Errorf("IsInRange(%v, %v, %v) = %v, expected %v",
					tt.testTime, start, end, result, tt.expected)
			}
		})
	}
}

func TestStartOfWeek_MonthBoundary(t *testing.T) {
	// Test cases where week crosses month boundary
	tests := []struct {
		name           string
		input          time.Time
		expectedMonday time.Time
	}{
		{
			name:           "February 1 (Thursday in 2024)",
			input:          makeTime(2024, time.February, 1, 10, 0, 0),
			expectedMonday: makeTime(2024, time.January, 29, 0, 0, 0),
		},
		{
			name:           "March 1 (Friday in 2024)",
			input:          makeTime(2024, time.March, 1, 10, 0, 0),
			expectedMonday: makeTime(2024, time.February, 26, 0, 0, 0),
		},
		{
			name:           "January 1 2024 (Monday)",
			input:          makeTime(2024, time.January, 1, 10, 0, 0),
			expectedMonday: makeTime(2024, time.January, 1, 0, 0, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StartOfWeek(tt.input)
			if !result.Equal(tt.expectedMonday) {
				t.Errorf("StartOfWeek(%v) = %v, expected %v", tt.input, result, tt.expectedMonday)
			}
		})
	}
}

func TestEndOfWeek_MonthBoundary(t *testing.T) {
	// Test cases where week crosses month boundary
	tests := []struct {
		name           string
		input          time.Time
		expectedSunday time.Time
	}{
		{
			name:           "January 29 2024 (Monday)",
			input:          makeTime(2024, time.January, 29, 10, 0, 0),
			expectedSunday: makeTime(2024, time.February, 4, 23, 59, 59),
		},
		{
			name:           "February 26 2024 (Monday)",
			input:          makeTime(2024, time.February, 26, 10, 0, 0),
			expectedSunday: makeTime(2024, time.March, 3, 23, 59, 59),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EndOfWeek(tt.input)

			// Check the date components
			if result.Year() != tt.expectedSunday.Year() || result.Month() != tt.expectedSunday.Month() || result.Day() != tt.expectedSunday.Day() {
				t.Errorf("EndOfWeek(%v) date = %v, expected %v", tt.input, result, tt.expectedSunday)
			}
		})
	}
}

func TestTimezonePreservation(t *testing.T) {
	// Ensure timezone is preserved through all operations
	localTime := makeTime(2024, time.January, 15, 14, 30, 0)
	originalLocation := localTime.Location()

	// Test StartOfDay preserves timezone
	result := StartOfDay(localTime)
	if result.Location() != originalLocation {
		t.Errorf("StartOfDay changed timezone from %v to %v", originalLocation, result.Location())
	}

	// Test EndOfDay preserves timezone
	result = EndOfDay(localTime)
	if result.Location() != originalLocation {
		t.Errorf("EndOfDay changed timezone from %v to %v", originalLocation, result.Location())
	}

	// Test StartOfWeek preserves timezone
	result = StartOfWeek(localTime)
	if result.Location() != originalLocation {
		t.Errorf("StartOfWeek changed timezone from %v to %v", originalLocation, result.Location())
	}

	// Test EndOfWeek preserves timezone
	result = EndOfWeek(localTime)
	if result.Location() != originalLocation {
		t.Errorf("EndOfWeek changed timezone from %v to %v", originalLocation, result.Location())
	}
}
