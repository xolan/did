package service

import (
	"fmt"
	"time"

	"github.com/xolan/did/internal/config"
	"github.com/xolan/did/internal/stats"
	"github.com/xolan/did/internal/storage"
	"github.com/xolan/did/internal/timeutil"
)

// StatsService provides statistics operations
type StatsService struct {
	storagePath string
	config      config.Config
}

// NewStatsService creates a new StatsService
func NewStatsService(storagePath string, cfg config.Config) *StatsService {
	return &StatsService{
		storagePath: storagePath,
		config:      cfg,
	}
}

// Weekly returns weekly statistics with comparison to previous week
func (s *StatsService) Weekly() (*StatsResult, error) {
	now := time.Now()

	// This week
	thisWeekStart := timeutil.StartOfWeekWithConfig(now, s.config.WeekStartDay)
	thisWeekEnd := timeutil.EndOfWeekWithConfig(now, s.config.WeekStartDay)

	// Last week
	lastWeekStart := thisWeekStart.AddDate(0, 0, -7)
	lastWeekEnd := timeutil.EndOfWeekWithConfig(lastWeekStart, s.config.WeekStartDay)

	return s.calculateStats(thisWeekStart, thisWeekEnd, lastWeekStart, lastWeekEnd, "this week", "week")
}

// Monthly returns monthly statistics with comparison to previous month
func (s *StatsService) Monthly() (*StatsResult, error) {
	// This month
	thisMonthStart, thisMonthEnd := timeutil.ThisMonth()

	// Last month
	lastMonthStart, lastMonthEnd := timeutil.LastMonth()

	return s.calculateStats(thisMonthStart, thisMonthEnd, lastMonthStart, lastMonthEnd, "this month", "month")
}

// ForDateRange returns statistics for a custom date range
func (s *StatsService) ForDateRange(spec DateRangeSpec) (*StatsResult, error) {
	start, end, period := s.resolveDateRange(spec)

	entries, err := storage.ReadActiveEntries(s.storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read entries: %w", err)
	}

	statistics := stats.CalculateStatistics(entries, start, end)
	projectStats := stats.CalculateProjectBreakdown(entries, start, end)
	tagStats := stats.CalculateTagBreakdown(entries, start, end)

	return &StatsResult{
		Statistics:   statistics,
		ProjectStats: projectStats,
		TagStats:     tagStats,
		Period:       period,
		Start:        start,
		End:          end,
	}, nil
}

// calculateStats calculates statistics for current and previous periods
func (s *StatsService) calculateStats(
	currentStart, currentEnd time.Time,
	previousStart, previousEnd time.Time,
	period, periodName string,
) (*StatsResult, error) {
	entries, err := storage.ReadActiveEntries(s.storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read entries: %w", err)
	}

	// Calculate current period stats
	currentStats := stats.CalculateStatistics(entries, currentStart, currentEnd)
	projectStats := stats.CalculateProjectBreakdown(entries, currentStart, currentEnd)
	tagStats := stats.CalculateTagBreakdown(entries, currentStart, currentEnd)

	// Calculate previous period stats for comparison
	previousStats := stats.CalculateStatistics(entries, previousStart, previousEnd)

	// Compare statistics
	diff := stats.CompareStatistics(currentStats, previousStats)
	comparison := stats.FormatComparison(diff, periodName)

	return &StatsResult{
		Statistics:    currentStats,
		ProjectStats:  projectStats,
		TagStats:      tagStats,
		Comparison:    comparison,
		Period:        period,
		Start:         currentStart,
		End:           currentEnd,
		PreviousStart: previousStart,
		PreviousEnd:   previousEnd,
	}, nil
}

// resolveDateRange converts a DateRangeSpec to concrete start/end times
func (s *StatsService) resolveDateRange(spec DateRangeSpec) (start, end time.Time, period string) {
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
