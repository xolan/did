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

func TestStartOfMonth(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "first day stays first day",
			input:    makeTime(2024, time.January, 1, 0, 0, 0),
			expected: makeTime(2024, time.January, 1, 0, 0, 0),
		},
		{
			name:     "middle of month goes to first",
			input:    makeTime(2024, time.January, 15, 14, 30, 0),
			expected: makeTime(2024, time.January, 1, 0, 0, 0),
		},
		{
			name:     "last day of month goes to first",
			input:    makeTime(2024, time.January, 31, 23, 59, 59),
			expected: makeTime(2024, time.January, 1, 0, 0, 0),
		},
		{
			name:     "February in leap year",
			input:    makeTime(2024, time.February, 15, 10, 0, 0),
			expected: makeTime(2024, time.February, 1, 0, 0, 0),
		},
		{
			name:     "February 29 in leap year",
			input:    makeTime(2024, time.February, 29, 18, 45, 30),
			expected: makeTime(2024, time.February, 1, 0, 0, 0),
		},
		{
			name:     "30-day month (April)",
			input:    makeTime(2024, time.April, 30, 12, 0, 0),
			expected: makeTime(2024, time.April, 1, 0, 0, 0),
		},
		{
			name:     "December",
			input:    makeTime(2024, time.December, 25, 9, 30, 0),
			expected: makeTime(2024, time.December, 1, 0, 0, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StartOfMonth(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("StartOfMonth(%v) = %v, expected %v", tt.input, result, tt.expected)
			}
			// Verify it's the first day of the month
			if result.Day() != 1 {
				t.Errorf("StartOfMonth(%v) day = %d, expected 1", tt.input, result.Day())
			}
			// Verify it's midnight
			if result.Hour() != 0 || result.Minute() != 0 || result.Second() != 0 || result.Nanosecond() != 0 {
				t.Errorf("StartOfMonth(%v) time components not all zero: got %02d:%02d:%02d.%09d",
					tt.input, result.Hour(), result.Minute(), result.Second(), result.Nanosecond())
			}
		})
	}
}

func TestEndOfMonth(t *testing.T) {
	tests := []struct {
		name        string
		input       time.Time
		expectedDay int // expected last day of month
	}{
		{
			name:        "January - 31 days",
			input:       makeTime(2024, time.January, 15, 12, 0, 0),
			expectedDay: 31,
		},
		{
			name:        "February leap year - 29 days",
			input:       makeTime(2024, time.February, 1, 0, 0, 0),
			expectedDay: 29,
		},
		{
			name:        "February non-leap year - 28 days",
			input:       makeTime(2023, time.February, 15, 10, 0, 0),
			expectedDay: 28,
		},
		{
			name:        "March - 31 days",
			input:       makeTime(2024, time.March, 20, 14, 30, 0),
			expectedDay: 31,
		},
		{
			name:        "April - 30 days",
			input:       makeTime(2024, time.April, 15, 9, 0, 0),
			expectedDay: 30,
		},
		{
			name:        "May - 31 days",
			input:       makeTime(2024, time.May, 31, 23, 59, 59),
			expectedDay: 31,
		},
		{
			name:        "June - 30 days",
			input:       makeTime(2024, time.June, 1, 0, 0, 0),
			expectedDay: 30,
		},
		{
			name:        "December - 31 days",
			input:       makeTime(2024, time.December, 25, 18, 0, 0),
			expectedDay: 31,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EndOfMonth(tt.input)

			// Should be same month and year
			if result.Year() != tt.input.Year() || result.Month() != tt.input.Month() {
				t.Errorf("EndOfMonth(%v) changed month/year: got %v", tt.input, result)
			}

			// Should be the expected last day
			if result.Day() != tt.expectedDay {
				t.Errorf("EndOfMonth(%v) day = %d, expected %d", tt.input, result.Day(), tt.expectedDay)
			}

			// Should be 23:59:59.999999999
			if result.Hour() != 23 || result.Minute() != 59 || result.Second() != 59 || result.Nanosecond() != 999999999 {
				t.Errorf("EndOfMonth(%v) = %02d:%02d:%02d.%09d, expected 23:59:59.999999999",
					tt.input, result.Hour(), result.Minute(), result.Second(), result.Nanosecond())
			}
		})
	}
}

func TestThisMonth(t *testing.T) {
	start, end := ThisMonth()
	now := time.Now()

	// Start should be first day of current month at midnight
	if start.Year() != now.Year() || start.Month() != now.Month() || start.Day() != 1 {
		t.Errorf("ThisMonth() start date mismatch: got %v, expected first of %v %d", start, now.Month(), now.Year())
	}
	if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
		t.Errorf("ThisMonth() start not midnight: got %02d:%02d:%02d", start.Hour(), start.Minute(), start.Second())
	}

	// End should be last day of current month at end of day
	if end.Year() != now.Year() || end.Month() != now.Month() {
		t.Errorf("ThisMonth() end month/year mismatch: got %v, expected %v %d", end, now.Month(), now.Year())
	}
	if end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 {
		t.Errorf("ThisMonth() end not end of day: got %02d:%02d:%02d", end.Hour(), end.Minute(), end.Second())
	}

	// Start should be before end
	if !start.Before(end) {
		t.Errorf("ThisMonth() start %v not before end %v", start, end)
	}

	// Verify start is first of month
	if start.Day() != 1 {
		t.Errorf("ThisMonth() start day = %d, expected 1", start.Day())
	}
}

func TestLastMonth(t *testing.T) {
	start, end := LastMonth()
	now := time.Now()
	expectedMonth := now.AddDate(0, -1, 0)

	// Start should be first day of last month at midnight
	if start.Year() != expectedMonth.Year() || start.Month() != expectedMonth.Month() || start.Day() != 1 {
		t.Errorf("LastMonth() start date mismatch: got %v, expected first of %v %d", start, expectedMonth.Month(), expectedMonth.Year())
	}
	if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
		t.Errorf("LastMonth() start not midnight: got %02d:%02d:%02d", start.Hour(), start.Minute(), start.Second())
	}

	// End should be last day of last month at end of day
	if end.Year() != expectedMonth.Year() || end.Month() != expectedMonth.Month() {
		t.Errorf("LastMonth() end month/year mismatch: got %v, expected %v %d", end, expectedMonth.Month(), expectedMonth.Year())
	}
	if end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 {
		t.Errorf("LastMonth() end not end of day: got %02d:%02d:%02d", end.Hour(), end.Minute(), end.Second())
	}

	// Start should be before end
	if !start.Before(end) {
		t.Errorf("LastMonth() start %v not before end %v", start, end)
	}

	// Verify start is first of month
	if start.Day() != 1 {
		t.Errorf("LastMonth() start day = %d, expected 1", start.Day())
	}

	// Last month end should be before this month start
	thisMonthStart, _ := ThisMonth()
	if !end.Before(thisMonthStart) {
		t.Errorf("LastMonth() end %v not before ThisMonth() start %v", end, thisMonthStart)
	}
}

func TestStartOfWeekWithConfig_Monday(t *testing.T) {
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
		{
			name:           "Sunday at midnight",
			input:          makeTime(2024, time.January, 21, 0, 0, 0), // Sunday
			expectedMonday: makeTime(2024, time.January, 15, 0, 0, 0),
		},
		{
			name:           "Sunday at end of day",
			input:          makeTime(2024, time.January, 21, 23, 59, 59), // Sunday
			expectedMonday: makeTime(2024, time.January, 15, 0, 0, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StartOfWeekWithConfig(tt.input, "monday")
			if !result.Equal(tt.expectedMonday) {
				t.Errorf("StartOfWeekWithConfig(%v [%s], monday) = %v [%s], expected %v [Monday]",
					tt.input, tt.input.Weekday(), result, result.Weekday(), tt.expectedMonday)
			}

			// Verify it's a Monday
			if result.Weekday() != time.Monday {
				t.Errorf("StartOfWeekWithConfig(%v, monday) weekday = %s, expected Monday", tt.input, result.Weekday())
			}

			// Verify it's midnight
			if result.Hour() != 0 || result.Minute() != 0 || result.Second() != 0 || result.Nanosecond() != 0 {
				t.Errorf("StartOfWeekWithConfig(%v, monday) not midnight: got %02d:%02d:%02d.%09d",
					tt.input, result.Hour(), result.Minute(), result.Second(), result.Nanosecond())
			}
		})
	}
}

func TestStartOfWeekWithConfig_Sunday(t *testing.T) {
	tests := []struct {
		name           string
		input          time.Time
		expectedSunday time.Time
	}{
		{
			name:           "Sunday stays Sunday",
			input:          makeTime(2024, time.January, 21, 10, 30, 0), // Sunday
			expectedSunday: makeTime(2024, time.January, 21, 0, 0, 0),
		},
		{
			name:           "Monday goes to previous Sunday",
			input:          makeTime(2024, time.January, 15, 14, 0, 0), // Monday
			expectedSunday: makeTime(2024, time.January, 14, 0, 0, 0),
		},
		{
			name:           "Tuesday goes to previous Sunday",
			input:          makeTime(2024, time.January, 16, 9, 15, 0), // Tuesday
			expectedSunday: makeTime(2024, time.January, 14, 0, 0, 0),
		},
		{
			name:           "Wednesday goes to previous Sunday",
			input:          makeTime(2024, time.January, 17, 16, 45, 0), // Wednesday
			expectedSunday: makeTime(2024, time.January, 14, 0, 0, 0),
		},
		{
			name:           "Thursday goes to previous Sunday",
			input:          makeTime(2024, time.January, 18, 11, 0, 0), // Thursday
			expectedSunday: makeTime(2024, time.January, 14, 0, 0, 0),
		},
		{
			name:           "Friday goes to previous Sunday",
			input:          makeTime(2024, time.January, 19, 8, 30, 0), // Friday
			expectedSunday: makeTime(2024, time.January, 14, 0, 0, 0),
		},
		{
			name:           "Saturday goes to previous Sunday",
			input:          makeTime(2024, time.January, 20, 20, 0, 0), // Saturday
			expectedSunday: makeTime(2024, time.January, 14, 0, 0, 0),
		},
		{
			name:           "Sunday at midnight",
			input:          makeTime(2024, time.January, 21, 0, 0, 0), // Sunday
			expectedSunday: makeTime(2024, time.January, 21, 0, 0, 0),
		},
		{
			name:           "Sunday at end of day",
			input:          makeTime(2024, time.January, 21, 23, 59, 59), // Sunday
			expectedSunday: makeTime(2024, time.January, 21, 0, 0, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StartOfWeekWithConfig(tt.input, "sunday")
			if !result.Equal(tt.expectedSunday) {
				t.Errorf("StartOfWeekWithConfig(%v [%s], sunday) = %v [%s], expected %v [Sunday]",
					tt.input, tt.input.Weekday(), result, result.Weekday(), tt.expectedSunday)
			}

			// Verify it's a Sunday
			if result.Weekday() != time.Sunday {
				t.Errorf("StartOfWeekWithConfig(%v, sunday) weekday = %s, expected Sunday", tt.input, result.Weekday())
			}

			// Verify it's midnight
			if result.Hour() != 0 || result.Minute() != 0 || result.Second() != 0 || result.Nanosecond() != 0 {
				t.Errorf("StartOfWeekWithConfig(%v, sunday) not midnight: got %02d:%02d:%02d.%09d",
					tt.input, result.Hour(), result.Minute(), result.Second(), result.Nanosecond())
			}
		})
	}
}

func TestStartOfWeekWithConfig_SundayEdgeCases(t *testing.T) {
	// Test the critical edge case: what happens on Sunday with different week start settings
	sundayDate := makeTime(2024, time.January, 21, 14, 30, 0) // Sunday Jan 21, 2024

	t.Run("Sunday with monday start goes to previous Monday", func(t *testing.T) {
		result := StartOfWeekWithConfig(sundayDate, "monday")
		expected := makeTime(2024, time.January, 15, 0, 0, 0) // Monday Jan 15

		if !result.Equal(expected) {
			t.Errorf("StartOfWeekWithConfig(Sunday, monday) = %v [%s], expected %v [Monday]",
				result, result.Weekday(), expected)
		}
		if result.Weekday() != time.Monday {
			t.Errorf("Expected Monday, got %s", result.Weekday())
		}
	})

	t.Run("Sunday with sunday start stays same Sunday", func(t *testing.T) {
		result := StartOfWeekWithConfig(sundayDate, "sunday")
		expected := makeTime(2024, time.January, 21, 0, 0, 0) // Same Sunday at midnight

		if !result.Equal(expected) {
			t.Errorf("StartOfWeekWithConfig(Sunday, sunday) = %v [%s], expected %v [Sunday]",
				result, result.Weekday(), expected)
		}
		if result.Weekday() != time.Sunday {
			t.Errorf("Expected Sunday, got %s", result.Weekday())
		}
	})
}

func TestStartOfWeekWithConfig_MondayEdgeCases(t *testing.T) {
	// Test the edge case: what happens on Monday with different week start settings
	mondayDate := makeTime(2024, time.January, 15, 14, 30, 0) // Monday Jan 15, 2024

	t.Run("Monday with monday start stays same Monday", func(t *testing.T) {
		result := StartOfWeekWithConfig(mondayDate, "monday")
		expected := makeTime(2024, time.January, 15, 0, 0, 0) // Same Monday at midnight

		if !result.Equal(expected) {
			t.Errorf("StartOfWeekWithConfig(Monday, monday) = %v [%s], expected %v [Monday]",
				result, result.Weekday(), expected)
		}
		if result.Weekday() != time.Monday {
			t.Errorf("Expected Monday, got %s", result.Weekday())
		}
	})

	t.Run("Monday with sunday start goes to previous Sunday", func(t *testing.T) {
		result := StartOfWeekWithConfig(mondayDate, "sunday")
		expected := makeTime(2024, time.January, 14, 0, 0, 0) // Previous Sunday Jan 14

		if !result.Equal(expected) {
			t.Errorf("StartOfWeekWithConfig(Monday, sunday) = %v [%s], expected %v [Sunday]",
				result, result.Weekday(), expected)
		}
		if result.Weekday() != time.Sunday {
			t.Errorf("Expected Sunday, got %s", result.Weekday())
		}
	})
}

func TestStartOfWeekWithConfig_MonthBoundary(t *testing.T) {
	tests := []struct {
		name           string
		input          time.Time
		weekStartDay   string
		expectedStart  time.Time
		expectedWeekday time.Weekday
	}{
		{
			name:           "Feb 1 with monday start crosses to January",
			input:          makeTime(2024, time.February, 1, 10, 0, 0), // Thursday Feb 1
			weekStartDay:   "monday",
			expectedStart:  makeTime(2024, time.January, 29, 0, 0, 0), // Monday Jan 29
			expectedWeekday: time.Monday,
		},
		{
			name:           "Feb 1 with sunday start stays in February",
			input:          makeTime(2024, time.February, 1, 10, 0, 0), // Thursday Feb 1
			weekStartDay:   "sunday",
			expectedStart:  makeTime(2024, time.January, 28, 0, 0, 0), // Sunday Jan 28
			expectedWeekday: time.Sunday,
		},
		{
			name:           "March 3 (Sunday) with sunday start",
			input:          makeTime(2024, time.March, 3, 15, 30, 0), // Sunday March 3
			weekStartDay:   "sunday",
			expectedStart:  makeTime(2024, time.March, 3, 0, 0, 0), // Same Sunday
			expectedWeekday: time.Sunday,
		},
		{
			name:           "March 3 (Sunday) with monday start crosses to February",
			input:          makeTime(2024, time.March, 3, 15, 30, 0), // Sunday March 3
			weekStartDay:   "monday",
			expectedStart:  makeTime(2024, time.February, 26, 0, 0, 0), // Monday Feb 26
			expectedWeekday: time.Monday,
		},
		{
			name:           "Jan 1 2024 (Monday) with monday start",
			input:          makeTime(2024, time.January, 1, 10, 0, 0),
			weekStartDay:   "monday",
			expectedStart:  makeTime(2024, time.January, 1, 0, 0, 0),
			expectedWeekday: time.Monday,
		},
		{
			name:           "Jan 1 2024 (Monday) with sunday start crosses to previous year",
			input:          makeTime(2024, time.January, 1, 10, 0, 0),
			weekStartDay:   "sunday",
			expectedStart:  makeTime(2023, time.December, 31, 0, 0, 0), // Sunday Dec 31, 2023
			expectedWeekday: time.Sunday,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StartOfWeekWithConfig(tt.input, tt.weekStartDay)
			if !result.Equal(tt.expectedStart) {
				t.Errorf("StartOfWeekWithConfig(%v, %s) = %v, expected %v",
					tt.input, tt.weekStartDay, result, tt.expectedStart)
			}
			if result.Weekday() != tt.expectedWeekday {
				t.Errorf("StartOfWeekWithConfig(%v, %s) weekday = %s, expected %s",
					tt.input, tt.weekStartDay, result.Weekday(), tt.expectedWeekday)
			}
		})
	}
}

func TestEndOfWeekWithConfig_Monday(t *testing.T) {
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
			result := EndOfWeekWithConfig(tt.input, "monday")

			// Check the date components
			if result.Year() != tt.expectedSunday.Year() || result.Month() != tt.expectedSunday.Month() || result.Day() != tt.expectedSunday.Day() {
				t.Errorf("EndOfWeekWithConfig(%v, monday) date = %v, expected %v", tt.input, result, tt.expectedSunday)
			}

			// Verify it's a Sunday
			if result.Weekday() != time.Sunday {
				t.Errorf("EndOfWeekWithConfig(%v, monday) weekday = %s, expected Sunday", tt.input, result.Weekday())
			}

			// Should be 23:59:59.999999999
			if result.Hour() != 23 || result.Minute() != 59 || result.Second() != 59 || result.Nanosecond() != 999999999 {
				t.Errorf("EndOfWeekWithConfig(%v, monday) = %02d:%02d:%02d.%09d, expected 23:59:59.999999999",
					tt.input, result.Hour(), result.Minute(), result.Second(), result.Nanosecond())
			}
		})
	}
}

func TestEndOfWeekWithConfig_Sunday(t *testing.T) {
	tests := []struct {
		name             string
		input            time.Time
		expectedSaturday time.Time
	}{
		{
			name:             "Sunday to Saturday",
			input:            makeTime(2024, time.January, 21, 10, 0, 0), // Sunday
			expectedSaturday: makeTime(2024, time.January, 27, 23, 59, 59),
		},
		{
			name:             "Monday to Saturday",
			input:            makeTime(2024, time.January, 15, 14, 30, 0), // Monday
			expectedSaturday: makeTime(2024, time.January, 20, 23, 59, 59),
		},
		{
			name:             "Wednesday to Saturday",
			input:            makeTime(2024, time.January, 17, 8, 0, 0), // Wednesday
			expectedSaturday: makeTime(2024, time.January, 20, 23, 59, 59),
		},
		{
			name:             "Saturday stays same Saturday",
			input:            makeTime(2024, time.January, 20, 16, 0, 0), // Saturday
			expectedSaturday: makeTime(2024, time.January, 20, 23, 59, 59),
		},
		{
			name:             "Friday to Saturday",
			input:            makeTime(2024, time.January, 19, 11, 30, 0), // Friday
			expectedSaturday: makeTime(2024, time.January, 20, 23, 59, 59),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EndOfWeekWithConfig(tt.input, "sunday")

			// Check the date components
			if result.Year() != tt.expectedSaturday.Year() || result.Month() != tt.expectedSaturday.Month() || result.Day() != tt.expectedSaturday.Day() {
				t.Errorf("EndOfWeekWithConfig(%v, sunday) date = %v, expected %v", tt.input, result, tt.expectedSaturday)
			}

			// Verify it's a Saturday
			if result.Weekday() != time.Saturday {
				t.Errorf("EndOfWeekWithConfig(%v, sunday) weekday = %s, expected Saturday", tt.input, result.Weekday())
			}

			// Should be 23:59:59.999999999
			if result.Hour() != 23 || result.Minute() != 59 || result.Second() != 59 || result.Nanosecond() != 999999999 {
				t.Errorf("EndOfWeekWithConfig(%v, sunday) = %02d:%02d:%02d.%09d, expected 23:59:59.999999999",
					tt.input, result.Hour(), result.Minute(), result.Second(), result.Nanosecond())
			}
		})
	}
}

func TestEndOfWeekWithConfig_WeekDuration(t *testing.T) {
	tests := []struct {
		name         string
		input        time.Time
		weekStartDay string
	}{
		{
			name:         "monday start week",
			input:        makeTime(2024, time.January, 17, 12, 0, 0),
			weekStartDay: "monday",
		},
		{
			name:         "sunday start week",
			input:        makeTime(2024, time.January, 17, 12, 0, 0),
			weekStartDay: "sunday",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := StartOfWeekWithConfig(tt.input, tt.weekStartDay)
			end := EndOfWeekWithConfig(tt.input, tt.weekStartDay)

			// Start should be before end
			if !start.Before(end) {
				t.Errorf("Week start %v not before end %v", start, end)
			}

			// Duration should be approximately 7 days
			duration := end.Sub(start)
			expectedDuration := 7*24*time.Hour - time.Nanosecond
			if duration != expectedDuration {
				t.Errorf("Week duration = %v, expected %v", duration, expectedDuration)
			}
		})
	}
}

func TestStartOfWeekWithConfig_DefaultsToMonday(t *testing.T) {
	// Test that non-"sunday" values default to monday behavior
	input := makeTime(2024, time.January, 17, 12, 0, 0) // Wednesday

	tests := []struct {
		name         string
		weekStartDay string
	}{
		{"empty string", ""},
		{"invalid value", "tuesday"},
		{"uppercase Monday", "Monday"},
		{"SUNDAY uppercase", "SUNDAY"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StartOfWeekWithConfig(input, tt.weekStartDay)
			expectedMonday := StartOfWeek(input)

			if !result.Equal(expectedMonday) {
				t.Errorf("StartOfWeekWithConfig(%v, %q) = %v [%s], expected %v [Monday] (should default to monday)",
					input, tt.weekStartDay, result, result.Weekday(), expectedMonday)
			}
		})
	}
}
