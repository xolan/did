package timeutil

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// ParseDate parses a date string in YYYY-MM-DD or DD/MM/YYYY format.
// Returns the parsed date at midnight (start of day) in local timezone.
// For ambiguous dates (like 05/06/2024), ISO format (YYYY-MM-DD) is preferred.
//
// Valid inputs:
//   - "2024-01-15" (ISO format)
//   - "15/01/2024" (European format)
//
// Invalid inputs return an error with suggested formats.
func ParseDate(input string) (time.Time, error) {
	if input == "" {
		return time.Time{}, fmt.Errorf("invalid date: empty string")
	}

	// Try ISO format first (YYYY-MM-DD) - preferred for ambiguous dates
	t, err := time.ParseInLocation("2006-01-02", input, time.Local)
	if err == nil {
		return StartOfDay(t), nil
	}

	// Try European format (DD/MM/YYYY)
	t, err = time.ParseInLocation("02/01/2006", input, time.Local)
	if err == nil {
		return StartOfDay(t), nil
	}

	// Neither format worked
	return time.Time{}, fmt.Errorf("invalid date format: expected YYYY-MM-DD or DD/MM/YYYY, got %s", input)
}

// ParseRelativeDays parses relative day expressions like "last N days".
// Returns the start and end times for the range.
// The range includes N complete days ending today (inclusive).
//
// Valid inputs:
//   - "last 7 days" (returns 7-day range ending today)
//   - "last 30 days" (returns 30-day range ending today)
//   - "last 1 day" (returns today only)
//
// Invalid inputs return an error with suggested format.
func ParseRelativeDays(input string) (start, end time.Time, err error) {
	if input == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid relative date: empty string")
	}

	// Match "last N days" or "last N day" (singular)
	// Use strict whitespace matching - single spaces only
	re := regexp.MustCompile(`^last\s(\d+)\sdays?$`)
	matches := re.FindStringSubmatch(input)

	if matches == nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid relative date format: expected 'last N days', got %s", input)
	}

	// Extract the number of days
	n, err := strconv.Atoi(matches[1])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid number in relative date: %s", matches[1])
	}

	if n <= 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid number of days: must be positive, got %d", n)
	}

	// Calculate the range: N days ending today (inclusive)
	// For "last 1 day": today only
	// For "last 7 days": 7 days ending today
	now := time.Now()
	endTime := EndOfDay(now)
	startTime := StartOfDay(now.AddDate(0, 0, -(n - 1)))

	return startTime, endTime, nil
}
