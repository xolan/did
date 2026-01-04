package timeutil

import (
	"fmt"
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
