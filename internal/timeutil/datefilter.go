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
	lastWeek := time.Now().AddDate(0, 0, -7)
	return StartOfWeek(lastWeek), EndOfWeek(lastWeek)
}

// IsInRange checks if the given time t falls within the range [start, end] (inclusive)
func IsInRange(t, start, end time.Time) bool {
	return (t.Equal(start) || t.After(start)) && (t.Equal(end) || t.Before(end))
}
