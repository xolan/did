package service

import (
	"fmt"
	"sort"
	"time"

	"github.com/xolan/did/internal/config"
	"github.com/xolan/did/internal/entry"
	"github.com/xolan/did/internal/filter"
	"github.com/xolan/did/internal/storage"
	"github.com/xolan/did/internal/timeutil"
)

// ReportService provides operations for generating reports
type ReportService struct {
	storagePath string
	config      config.Config
}

// NewReportService creates a new ReportService
func NewReportService(storagePath string, cfg config.Config) *ReportService {
	return &ReportService{
		storagePath: storagePath,
		config:      cfg,
	}
}

// ByProject generates a report for a specific project
func (s *ReportService) ByProject(project string, dateRange DateRangeSpec) (*ReportData, error) {
	start, end, period := s.resolveDateRange(dateRange)

	entries, err := s.loadActiveEntries()
	if err != nil {
		return nil, err
	}

	// Filter by date range and project
	f := filter.NewFilter("", project, nil)
	var filtered []entry.Entry
	for _, e := range entries {
		if timeutil.IsInRange(e.Timestamp, start, end) && f.Matches(e) {
			filtered = append(filtered, e)
		}
	}

	// Calculate totals
	totalMinutes := 0
	for _, e := range filtered {
		totalMinutes += e.DurationMinutes
	}

	return &ReportData{
		Groups: []GroupData{{
			Name:         project,
			TotalMinutes: totalMinutes,
			EntryCount:   len(filtered),
		}},
		TotalMinutes: totalMinutes,
		EntryCount:   len(filtered),
		Period:       period,
		Start:        start,
		End:          end,
	}, nil
}

// ByTags generates a report for entries matching specific tags
func (s *ReportService) ByTags(tags []string, dateRange DateRangeSpec) (*ReportData, error) {
	start, end, period := s.resolveDateRange(dateRange)

	entries, err := s.loadActiveEntries()
	if err != nil {
		return nil, err
	}

	// Filter by date range and tags
	f := filter.NewFilter("", "", tags)
	var filtered []entry.Entry
	for _, e := range entries {
		if timeutil.IsInRange(e.Timestamp, start, end) && f.Matches(e) {
			filtered = append(filtered, e)
		}
	}

	// Calculate totals
	totalMinutes := 0
	for _, e := range filtered {
		totalMinutes += e.DurationMinutes
	}

	return &ReportData{
		TotalMinutes: totalMinutes,
		EntryCount:   len(filtered),
		Period:       period,
		Start:        start,
		End:          end,
	}, nil
}

// GroupByProject generates a report grouped by project
func (s *ReportService) GroupByProject(dateRange DateRangeSpec) (*ReportData, error) {
	start, end, period := s.resolveDateRange(dateRange)

	entries, err := s.loadActiveEntries()
	if err != nil {
		return nil, err
	}

	// Filter by date range
	var filtered []entry.Entry
	for _, e := range entries {
		if timeutil.IsInRange(e.Timestamp, start, end) {
			filtered = append(filtered, e)
		}
	}

	// Group by project
	projectMap := make(map[string]*GroupData)
	totalMinutes := 0

	for _, e := range filtered {
		projectName := e.Project
		if projectName == "" {
			projectName = "(no project)"
		}

		if _, exists := projectMap[projectName]; !exists {
			projectMap[projectName] = &GroupData{Name: projectName}
		}

		projectMap[projectName].TotalMinutes += e.DurationMinutes
		projectMap[projectName].EntryCount++
		totalMinutes += e.DurationMinutes
	}

	// Convert to slice and sort by total minutes (descending)
	var groups []GroupData
	for _, group := range projectMap {
		groups = append(groups, *group)
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].TotalMinutes > groups[j].TotalMinutes
	})

	return &ReportData{
		Groups:       groups,
		TotalMinutes: totalMinutes,
		EntryCount:   len(filtered),
		Period:       period,
		Start:        start,
		End:          end,
	}, nil
}

// GroupByTag generates a report grouped by tag
func (s *ReportService) GroupByTag(dateRange DateRangeSpec) (*ReportData, error) {
	start, end, period := s.resolveDateRange(dateRange)

	entries, err := s.loadActiveEntries()
	if err != nil {
		return nil, err
	}

	// Filter by date range
	var filtered []entry.Entry
	for _, e := range entries {
		if timeutil.IsInRange(e.Timestamp, start, end) {
			filtered = append(filtered, e)
		}
	}

	// Group by tag (entries with multiple tags contribute to each)
	tagMap := make(map[string]*GroupData)
	totalMinutes := 0

	for _, e := range filtered {
		if len(e.Tags) == 0 {
			tagName := "(no tags)"
			if _, exists := tagMap[tagName]; !exists {
				tagMap[tagName] = &GroupData{Name: tagName}
			}
			tagMap[tagName].TotalMinutes += e.DurationMinutes
			tagMap[tagName].EntryCount++
		} else {
			for _, tag := range e.Tags {
				if _, exists := tagMap[tag]; !exists {
					tagMap[tag] = &GroupData{Name: tag}
				}
				tagMap[tag].TotalMinutes += e.DurationMinutes
				tagMap[tag].EntryCount++
			}
		}
		totalMinutes += e.DurationMinutes
	}

	// Convert to slice and sort by total minutes (descending)
	var groups []GroupData
	for _, group := range tagMap {
		groups = append(groups, *group)
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].TotalMinutes > groups[j].TotalMinutes
	})

	return &ReportData{
		Groups:       groups,
		TotalMinutes: totalMinutes,
		EntryCount:   len(filtered),
		Period:       period,
		Start:        start,
		End:          end,
	}, nil
}

// loadActiveEntries loads all active (non-deleted) entries
func (s *ReportService) loadActiveEntries() ([]entry.Entry, error) {
	entries, err := storage.ReadActiveEntries(s.storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read entries: %w", err)
	}
	return entries, nil
}

// resolveDateRange converts a DateRangeSpec to concrete start/end times
func (s *ReportService) resolveDateRange(spec DateRangeSpec) (start, end time.Time, period string) {
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
