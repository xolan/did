// Package service provides the business logic layer for the did application.
// It wraps the underlying storage, timer, config, and stats packages,
// providing a clean API for both CLI and TUI frontends.
package service

import (
	"time"

	"github.com/xolan/did/internal/entry"
	"github.com/xolan/did/internal/stats"
	"github.com/xolan/did/internal/storage"
	"github.com/xolan/did/internal/timer"
)

// DateRange represents a predefined or custom date range for filtering entries
type DateRange int

const (
	DateRangeToday DateRange = iota
	DateRangeYesterday
	DateRangeThisWeek
	DateRangePrevWeek
	DateRangeThisMonth
	DateRangePrevMonth
	DateRangeLast // Last N days (requires LastDays field)
	DateRangeCustom
)

// DateRangeSpec specifies a date range for filtering entries
type DateRangeSpec struct {
	Type     DateRange
	LastDays int       // Used when Type is DateRangeLast
	From     time.Time // Used when Type is DateRangeCustom
	To       time.Time // Used when Type is DateRangeCustom
}

// IndexedEntry represents an entry with its display index and storage index
type IndexedEntry struct {
	Entry        entry.Entry
	ActiveIndex  int // 1-based user-facing index (among active entries)
	StorageIndex int // 0-based index in storage (includes deleted entries)
}

// ListResult contains the results of listing entries
type ListResult struct {
	Entries  []IndexedEntry
	Warnings []storage.ParseWarning
	Period   string    // Human-readable period description
	Start    time.Time // Start of the date range
	End      time.Time // End of the date range
	Total    int       // Total duration in minutes
}

// TimerStatus represents the current state of the timer
type TimerStatus struct {
	Running     bool
	State       *timer.TimerState
	ElapsedTime time.Duration
}

// StatsResult contains statistics for a time period
type StatsResult struct {
	Statistics    stats.Statistics
	ProjectStats  []stats.ProjectBreakdown
	TagStats      []stats.TagBreakdown
	Comparison    string // Comparison with previous period (e.g., "up 2h from last week")
	Period        string // Human-readable period description
	Start         time.Time
	End           time.Time
	PreviousStart time.Time
	PreviousEnd   time.Time
}

// ReportData contains report data grouped by project or tag
type ReportData struct {
	Groups       []GroupData
	TotalMinutes int
	EntryCount   int
	Period       string
	Start        time.Time
	End          time.Time
}

// GroupData represents a single group in a report (project or tag)
type GroupData struct {
	Name         string
	TotalMinutes int
	EntryCount   int
}

// SearchResult contains search results
type SearchResult struct {
	Entries  []IndexedEntry
	Warnings []storage.ParseWarning
	Query    string // The search query used
	Total    int    // Total matching entries
}
