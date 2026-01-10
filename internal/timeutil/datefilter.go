package timeutil

import "time"

// StartOfDay returns midnight (00:00:00) of the given day in the same timezone
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay returns the last nanosecond of the given day (23:59:59.999999999)
func EndOfDay(t time.Time) time.Time {
	return StartOfDay(t).Add(24*time.Hour - time.Nanosecond)
}

// StartOfWeek returns Monday 00:00:00 of the week containing the given time (ISO standard)
// Handles the Sunday edge case where Go's Weekday() returns 0
func StartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 { // Sunday
		weekday = 7
	}
	return StartOfDay(t).AddDate(0, 0, -(weekday - 1))
}

// EndOfWeek returns Sunday 23:59:59.999999999 of the week containing the given time
func EndOfWeek(t time.Time) time.Time {
	return StartOfWeek(t).AddDate(0, 0, 7).Add(-time.Nanosecond)
}

// StartOfWeekWithConfig returns the start of week (00:00:00) based on the configured week start day
// weekStartDay should be "monday" or "sunday"
// For monday: returns Monday 00:00:00 of the week containing the given time (ISO standard)
// For sunday: returns Sunday 00:00:00 of the week containing the given time
func StartOfWeekWithConfig(t time.Time, weekStartDay string) time.Time {
	if weekStartDay == "sunday" {
		weekday := int(t.Weekday()) // Sunday = 0, Monday = 1, ..., Saturday = 6
		return StartOfDay(t).AddDate(0, 0, -weekday)
	}
	// Default to monday (ISO standard)
	return StartOfWeek(t)
}

// EndOfWeekWithConfig returns the end of week (23:59:59.999999999) based on the configured week start day
// weekStartDay should be "monday" or "sunday"
// For monday: returns Sunday 23:59:59.999999999 of the week
// For sunday: returns Saturday 23:59:59.999999999 of the week
func EndOfWeekWithConfig(t time.Time, weekStartDay string) time.Time {
	return StartOfWeekWithConfig(t, weekStartDay).AddDate(0, 0, 7).Add(-time.Nanosecond)
}

// Today returns the start and end times for today
func Today() (start, end time.Time) {
	now := time.Now()
	return StartOfDay(now), EndOfDay(now)
}

// Yesterday returns the start and end times for yesterday
func Yesterday() (start, end time.Time) {
	yesterday := time.Now().AddDate(0, 0, -1)
	return StartOfDay(yesterday), EndOfDay(yesterday)
}

// ThisWeek returns the start and end times for the current week (Monday-Sunday)
func ThisWeek() (start, end time.Time) {
	now := time.Now()
	return StartOfWeek(now), EndOfWeek(now)
}

// LastWeek returns the start and end times for the previous week (Monday-Sunday)
func LastWeek() (start, end time.Time) {
	thisWeekStart, _ := ThisWeek()
	lastWeekStart := thisWeekStart.AddDate(0, 0, -7)
	return lastWeekStart, EndOfWeek(lastWeekStart)
}

// StartOfMonth returns the first day of the month at 00:00:00 in the same timezone
func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth returns the last nanosecond of the last day of the month (23:59:59.999999999)
func EndOfMonth(t time.Time) time.Time {
	// Get the first day of next month, then subtract one nanosecond
	// This automatically handles different month lengths (28, 29, 30, 31 days)
	return StartOfMonth(t).AddDate(0, 1, 0).Add(-time.Nanosecond)
}

// ThisMonth returns the start and end times for the current month
func ThisMonth() (start, end time.Time) {
	now := time.Now()
	return StartOfMonth(now), EndOfMonth(now)
}

// LastMonth returns the start and end times for the previous month
func LastMonth() (start, end time.Time) {
	lastMonth := time.Now().AddDate(0, -1, 0)
	return StartOfMonth(lastMonth), EndOfMonth(lastMonth)
}

// IsInRange checks if the given time t falls within the range [start, end] (inclusive)
func IsInRange(t, start, end time.Time) bool {
	return (t.Equal(start) || t.After(start)) && (t.Equal(end) || t.Before(end))
}

// LoadTimezone loads a timezone by name. Returns Local if name is empty or "Local".
// Returns an error if the timezone name is invalid.
func LoadTimezone(tz string) (*time.Location, error) {
	if tz == "" || tz == "Local" {
		return time.Local, nil
	}
	return time.LoadLocation(tz)
}

// NowIn returns the current time in the specified timezone.
// Falls back to Local if timezone is invalid or empty.
func NowIn(tz string) time.Time {
	loc, err := LoadTimezone(tz)
	if err != nil {
		return time.Now()
	}
	return time.Now().In(loc)
}

// InTimezone converts a time to the specified timezone.
// Returns the original time if timezone is invalid or empty.
func InTimezone(t time.Time, tz string) time.Time {
	loc, err := LoadTimezone(tz)
	if err != nil {
		return t
	}
	return t.In(loc)
}
