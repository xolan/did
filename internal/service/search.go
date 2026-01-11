package service

import (
	"fmt"
	"sort"
	"time"

	"github.com/xolan/did/internal/config"
	"github.com/xolan/did/internal/filter"
	"github.com/xolan/did/internal/storage"
	"github.com/xolan/did/internal/timeutil"
)

// SearchService provides search operations for entries
type SearchService struct {
	storagePath string
	config      config.Config
}

// NewSearchService creates a new SearchService
func NewSearchService(storagePath string, cfg config.Config) *SearchService {
	return &SearchService{
		storagePath: storagePath,
		config:      cfg,
	}
}

// Search searches entries by keyword and optional filters
func (s *SearchService) Search(keyword string, dateRange *DateRangeSpec, f *filter.Filter) (*SearchResult, error) {
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

	// Apply date range filter if specified
	var filtered []IndexedEntry
	if dateRange != nil {
		start, end, _ := s.resolveDateRange(*dateRange)
		for _, ie := range activeEntries {
			if timeutil.IsInRange(ie.Entry.Timestamp, start, end) {
				filtered = append(filtered, ie)
			}
		}
	} else {
		filtered = activeEntries
	}

	// Apply keyword and project/tag filter
	searchFilter := filter.NewFilter(keyword, "", nil)
	if f != nil {
		searchFilter = filter.NewFilter(keyword, f.Project, f.Tags)
	}

	if !searchFilter.IsEmpty() {
		var searchFiltered []IndexedEntry
		for _, ie := range filtered {
			if searchFilter.Matches(ie.Entry) {
				searchFiltered = append(searchFiltered, ie)
			}
		}
		filtered = searchFiltered
	}

	// Sort by timestamp (most recent first for search results)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Entry.Timestamp.After(filtered[j].Entry.Timestamp)
	})

	return &SearchResult{
		Entries:  filtered,
		Warnings: result.Warnings,
		Query:    keyword,
		Total:    len(filtered),
	}, nil
}

// resolveDateRange converts a DateRangeSpec to concrete start/end times
func (s *SearchService) resolveDateRange(spec DateRangeSpec) (start, end time.Time, period string) {
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
		// No filter - use very wide range
		start = time.Time{}
		end = time.Now().AddDate(100, 0, 0)
		period = "all time"
	}

	return start, end, period
}
