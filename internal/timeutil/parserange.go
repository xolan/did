package timeutil

import (
	"fmt"
	"time"
)

// ParseDateRangeFlags parses date range flags and returns start/end times.
// If lastDays > 0, it takes precedence over from/to.
// Returns an error if both lastDays and from/to are specified.
func ParseDateRangeFlags(fromStr, toStr string, lastDays int) (start, end time.Time, err error) {
	if lastDays > 0 && (fromStr != "" || toStr != "") {
		return time.Time{}, time.Time{}, fmt.Errorf("cannot use --last with --from or --to")
	}

	if lastDays > 0 {
		now := time.Now()
		end = EndOfDay(now)
		start = StartOfDay(now.AddDate(0, 0, -(lastDays - 1)))
		return start, end, nil
	}

	if fromStr != "" {
		start, err = ParseDate(fromStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid --from date: %w", err)
		}
	}

	if toStr != "" {
		toDate, err := ParseDate(toStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid --to date: %w", err)
		}
		end = EndOfDay(toDate)
	} else {
		end = EndOfDay(time.Now())
	}

	if !start.IsZero() && start.After(end) {
		return time.Time{}, time.Time{}, fmt.Errorf("--from date (%s) is after --to date (%s)",
			start.Format("2006-01-02"), end.Format("2006-01-02"))
	}

	return start, end, nil
}
