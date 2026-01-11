package service

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/xolan/did/internal/config"
	"github.com/xolan/did/internal/entry"
	"github.com/xolan/did/internal/filter"
	"github.com/xolan/did/internal/storage"
	"github.com/xolan/did/internal/timeutil"
)

// Common errors for the entry service
var (
	ErrMissingDuration    = errors.New("missing 'for <duration>' in input")
	ErrEmptyDescription   = errors.New("description cannot be empty")
	ErrInvalidIndex       = errors.New("invalid entry index")
	ErrIndexOutOfRange    = errors.New("index out of range")
	ErrNoEntries          = errors.New("no entries found")
	ErrNoDeletedEntries   = errors.New("no deleted entries to restore")
	ErrNoChangesSpecified = errors.New("at least one change must be specified")
)

// EntryService provides operations for managing time tracking entries
type EntryService struct {
	storagePath string
	config      config.Config
}

// NewEntryService creates a new EntryService
func NewEntryService(storagePath string, cfg config.Config) *EntryService {
	return &EntryService{
		storagePath: storagePath,
		config:      cfg,
	}
}

// Create creates a new time tracking entry from a raw input string.
// Input format: "<description> for <duration>" (e.g., "fix bug @acme for 2h")
func (s *EntryService) Create(rawInput string) (*entry.Entry, error) {
	// Parse the input: expected format "<description> for <duration>"
	lastForIdx := strings.LastIndex(strings.ToLower(rawInput), " for ")
	if lastForIdx == -1 {
		return nil, ErrMissingDuration
	}

	description := strings.TrimSpace(rawInput[:lastForIdx])
	durationStr := strings.TrimSpace(rawInput[lastForIdx+5:]) // +5 for " for "

	if description == "" {
		return nil, ErrEmptyDescription
	}

	// Parse project and tags from description
	cleanDesc, project, tags := entry.ParseProjectAndTags(description)

	// Check that cleaned description is not empty
	if cleanDesc == "" {
		return nil, ErrEmptyDescription
	}

	// Parse the duration
	minutes, err := entry.ParseDuration(durationStr)
	if err != nil {
		return nil, fmt.Errorf("invalid duration '%s': %w", durationStr, err)
	}

	// Create the entry
	e := entry.Entry{
		Timestamp:       time.Now(),
		Description:     cleanDesc,
		DurationMinutes: minutes,
		RawInput:        rawInput,
		Project:         project,
		Tags:            tags,
	}

	// Append the entry to storage
	if err := storage.AppendEntry(s.storagePath, e); err != nil {
		return nil, fmt.Errorf("failed to save entry: %w", err)
	}

	return &e, nil
}

// CreateFromParts creates a new entry from individual components
func (s *EntryService) CreateFromParts(description string, durationMinutes int, project string, tags []string) (*entry.Entry, error) {
	if description == "" {
		return nil, ErrEmptyDescription
	}

	if durationMinutes <= 0 || durationMinutes > entry.MaxDurationMinutes {
		return nil, fmt.Errorf("invalid duration: must be 1-%d minutes", entry.MaxDurationMinutes)
	}

	// Create the entry
	e := entry.Entry{
		Timestamp:       time.Now(),
		Description:     description,
		DurationMinutes: durationMinutes,
		RawInput:        fmt.Sprintf("%s for %dm", description, durationMinutes),
		Project:         project,
		Tags:            tags,
	}

	// Append the entry to storage
	if err := storage.AppendEntry(s.storagePath, e); err != nil {
		return nil, fmt.Errorf("failed to save entry: %w", err)
	}

	return &e, nil
}

// List returns entries for the specified date range and filter
func (s *EntryService) List(dateRange DateRangeSpec, f *filter.Filter) (*ListResult, error) {
	// Get the time range
	start, end, period := s.resolveDateRange(dateRange)

	// Read all entries with warnings
	result, err := storage.ReadEntriesWithWarnings(s.storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read entries: %w", err)
	}

	// Build indexed entries (only active ones)
	var activeEntries []IndexedEntry
	activeIdx := 0
	for i, e := range result.Entries {
		if e.DeletedAt == nil {
			activeIdx++
			activeEntries = append(activeEntries, IndexedEntry{
				Entry:        e,
				ActiveIndex:  activeIdx,
				StorageIndex: i,
			})
		}
	}

	// Filter by time range
	var filtered []IndexedEntry
	for _, ie := range activeEntries {
		if timeutil.IsInRange(ie.Entry.Timestamp, start, end) {
			filtered = append(filtered, ie)
		}
	}

	// Apply project/tag filter
	if f != nil && !f.IsEmpty() {
		var projectTagFiltered []IndexedEntry
		for _, ie := range filtered {
			if f.Matches(ie.Entry) {
				projectTagFiltered = append(projectTagFiltered, ie)
			}
		}
		filtered = projectTagFiltered
	}

	// Sort by timestamp
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Entry.Timestamp.Before(filtered[j].Entry.Timestamp)
	})

	// Calculate total duration
	totalMinutes := 0
	for _, ie := range filtered {
		totalMinutes += ie.Entry.DurationMinutes
	}

	return &ListResult{
		Entries:  filtered,
		Warnings: result.Warnings,
		Period:   period,
		Start:    start,
		End:      end,
		Total:    totalMinutes,
	}, nil
}

// Edit updates an entry at the given user index (1-based)
func (s *EntryService) Edit(userIndex int, newDescription, newDuration string) (*entry.Entry, error) {
	if newDescription == "" && newDuration == "" {
		return nil, ErrNoChangesSpecified
	}

	if userIndex < 1 {
		return nil, ErrInvalidIndex
	}

	// Read all entries
	result, err := storage.ReadEntriesWithWarnings(s.storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read entries: %w", err)
	}

	// Build active entries mapping
	var activeEntries []entry.Entry
	var storageIndices []int
	for i, e := range result.Entries {
		if e.DeletedAt == nil {
			activeEntries = append(activeEntries, e)
			storageIndices = append(storageIndices, i)
		}
	}

	if len(activeEntries) == 0 {
		return nil, ErrNoEntries
	}

	// Convert 1-based to 0-based index
	activeIndex := userIndex - 1
	if activeIndex < 0 || activeIndex >= len(activeEntries) {
		return nil, fmt.Errorf("%w: valid range is 1-%d", ErrIndexOutOfRange, len(activeEntries))
	}

	e := activeEntries[activeIndex]
	storageIndex := storageIndices[activeIndex]

	// Update description if provided
	if newDescription != "" {
		cleanDesc, project, tags := entry.ParseProjectAndTags(newDescription)
		if cleanDesc == "" {
			return nil, ErrEmptyDescription
		}
		e.Description = cleanDesc
		e.Project = project
		e.Tags = tags
	}

	// Update duration if provided
	if newDuration != "" {
		minutes, err := entry.ParseDuration(newDuration)
		if err != nil {
			return nil, fmt.Errorf("invalid duration '%s': %w", newDuration, err)
		}
		e.DurationMinutes = minutes
	}

	// Update RawInput
	e.RawInput = s.buildRawInput(e)

	// Save the updated entry
	if err := storage.UpdateEntry(s.storagePath, storageIndex, e); err != nil {
		return nil, fmt.Errorf("failed to save entry: %w", err)
	}

	return &e, nil
}

// Delete soft-deletes an entry at the given user index (1-based)
func (s *EntryService) Delete(userIndex int) (*entry.Entry, error) {
	if userIndex < 1 {
		return nil, ErrInvalidIndex
	}

	// Read all entries
	allEntries, err := storage.ReadEntries(s.storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read entries: %w", err)
	}

	// Build active entries mapping
	var activeEntries []entry.Entry
	var storageIndices []int
	for i, e := range allEntries {
		if e.DeletedAt == nil {
			activeEntries = append(activeEntries, e)
			storageIndices = append(storageIndices, i)
		}
	}

	if len(activeEntries) == 0 {
		return nil, ErrNoEntries
	}

	// Convert 1-based to 0-based index
	activeIndex := userIndex - 1
	if activeIndex < 0 || activeIndex >= len(activeEntries) {
		return nil, fmt.Errorf("%w: valid range is 1-%d", ErrIndexOutOfRange, len(activeEntries))
	}

	storageIndex := storageIndices[activeIndex]

	// Soft delete the entry
	deletedEntry, err := storage.SoftDeleteEntry(s.storagePath, storageIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to delete entry: %w", err)
	}

	// Clean up old deleted entries (>7 days old)
	_, _ = storage.CleanupOldDeleted(s.storagePath)

	return &deletedEntry, nil
}

// Restore restores the most recently deleted entry
func (s *EntryService) Restore() (*entry.Entry, error) {
	// Find the most recently deleted entry
	_, index, err := storage.GetMostRecentlyDeleted(s.storagePath)
	if err != nil {
		return nil, ErrNoDeletedEntries
	}

	// Restore the entry
	restoredEntry, err := storage.RestoreEntry(s.storagePath, index)
	if err != nil {
		return nil, fmt.Errorf("failed to restore entry: %w", err)
	}

	return &restoredEntry, nil
}

// GetByIndex returns an entry at the given user index (1-based)
func (s *EntryService) GetByIndex(userIndex int) (*IndexedEntry, error) {
	if userIndex < 1 {
		return nil, ErrInvalidIndex
	}

	result, err := storage.ReadEntriesWithWarnings(s.storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read entries: %w", err)
	}

	// Build active entries mapping
	activeIdx := 0
	for i, e := range result.Entries {
		if e.DeletedAt == nil {
			activeIdx++
			if activeIdx == userIndex {
				return &IndexedEntry{
					Entry:        e,
					ActiveIndex:  activeIdx,
					StorageIndex: i,
				}, nil
			}
		}
	}

	return nil, ErrIndexOutOfRange
}

// GetActiveCount returns the number of active (non-deleted) entries
func (s *EntryService) GetActiveCount() (int, error) {
	entries, err := storage.ReadActiveEntries(s.storagePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read entries: %w", err)
	}
	return len(entries), nil
}

// resolveDateRange converts a DateRangeSpec to concrete start/end times and a period description
func (s *EntryService) resolveDateRange(spec DateRangeSpec) (start, end time.Time, period string) {
	now := time.Now()

	switch spec.Type {
	case DateRangeToday:
		start, end = timeutil.Today()
		period = "today"
	case DateRangeYesterday:
		start, end = timeutil.Yesterday()
		period = "yesterday"
	case DateRangeThisWeek:
		start = timeutil.StartOfWeekWithConfig(now, s.config.WeekStartDay)
		end = timeutil.EndOfWeekWithConfig(now, s.config.WeekStartDay)
		period = "this week"
	case DateRangePrevWeek:
		thisWeekStart := timeutil.StartOfWeekWithConfig(now, s.config.WeekStartDay)
		start = thisWeekStart.AddDate(0, 0, -7)
		end = timeutil.EndOfWeekWithConfig(start, s.config.WeekStartDay)
		period = "last week"
	case DateRangeThisMonth:
		start, end = timeutil.ThisMonth()
		period = "this month"
	case DateRangePrevMonth:
		start, end = timeutil.LastMonth()
		period = "last month"
	case DateRangeLast:
		end = timeutil.EndOfDay(now)
		start = timeutil.StartOfDay(now.AddDate(0, 0, -(spec.LastDays - 1)))
		period = fmt.Sprintf("last %d days", spec.LastDays)
	case DateRangeCustom:
		start = spec.From
		end = spec.To
		period = formatDateRangeForDisplay(start, end)
	default:
		start, end = timeutil.Today()
		period = "today"
	}

	return start, end, period
}

// buildRawInput reconstructs the raw input string from entry fields
func (s *EntryService) buildRawInput(e entry.Entry) string {
	desc := e.Description
	if e.Project != "" {
		desc += " @" + e.Project
	}
	for _, tag := range e.Tags {
		desc += " #" + tag
	}
	return fmt.Sprintf("%s for %s", desc, formatDurationSimple(e.DurationMinutes))
}

// formatDateRangeForDisplay formats a date range for human-readable display
func formatDateRangeForDisplay(start, end time.Time) string {
	if start.Format("2006-01-02") == end.Format("2006-01-02") {
		return start.Format("Mon, Jan 2, 2006")
	}
	if start.Year() == end.Year() {
		return fmt.Sprintf("%s - %s", start.Format("Jan 2"), end.Format("Jan 2, 2006"))
	}
	return fmt.Sprintf("%s - %s", start.Format("Jan 2, 2006"), end.Format("Jan 2, 2006"))
}

// formatDurationSimple formats minutes as a simple duration string
func formatDurationSimple(minutes int) string {
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	hours := minutes / 60
	mins := minutes % 60
	if mins == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh%dm", hours, mins)
}
