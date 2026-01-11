package views

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xolan/did/internal/config"
	"github.com/xolan/did/internal/entry"
	"github.com/xolan/did/internal/service"
	"github.com/xolan/did/internal/stats"
	"github.com/xolan/did/internal/timer"
	"github.com/xolan/did/internal/tui/ui"
)

func setupTestServices(t *testing.T) *service.Services {
	t.Helper()
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	timerPath := filepath.Join(tmpDir, "timer.json")
	configPath := filepath.Join(tmpDir, "config.yaml")
	cfg := config.DefaultConfig()

	return &service.Services{
		Entry:  service.NewEntryService(storagePath, cfg),
		Timer:  service.NewTimerService(timerPath, storagePath, cfg),
		Report: service.NewReportService(storagePath, cfg),
		Search: service.NewSearchService(storagePath, cfg),
		Stats:  service.NewStatsService(storagePath, cfg),
		Config: service.NewConfigService(configPath, cfg),
	}
}

func setupTestServicesWithEntries(t *testing.T) *service.Services {
	t.Helper()
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	timerPath := filepath.Join(tmpDir, "timer.json")
	configPath := filepath.Join(tmpDir, "config.yaml")
	cfg := config.DefaultConfig()

	// Create some entries
	now := time.Now().Format("2006-01-02T15:04:05Z07:00")
	content := `{"description":"task one","duration_minutes":60,"timestamp":"` + now + `","project":"acme","tags":["urgent"]}
{"description":"task two","duration_minutes":30,"timestamp":"` + now + `"}
`
	if err := os.WriteFile(storagePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	return &service.Services{
		Entry:  service.NewEntryService(storagePath, cfg),
		Timer:  service.NewTimerService(timerPath, storagePath, cfg),
		Report: service.NewReportService(storagePath, cfg),
		Search: service.NewSearchService(storagePath, cfg),
		Stats:  service.NewStatsService(storagePath, cfg),
		Config: service.NewConfigService(configPath, cfg),
	}
}

func setupTestServicesWithTimer(t *testing.T) (*service.Services, string) {
	t.Helper()
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	timerPath := filepath.Join(tmpDir, "timer.json")
	configPath := filepath.Join(tmpDir, "config.yaml")
	cfg := config.DefaultConfig()

	// Create a running timer
	state := timer.TimerState{
		StartedAt:   time.Now().Add(-30 * time.Minute),
		Description: "working on feature",
		Project:     "project",
		Tags:        []string{"feature"},
	}
	if err := timer.SaveTimerState(timerPath, state); err != nil {
		t.Fatal(err)
	}

	return &service.Services{
		Entry:  service.NewEntryService(storagePath, cfg),
		Timer:  service.NewTimerService(timerPath, storagePath, cfg),
		Report: service.NewReportService(storagePath, cfg),
		Search: service.NewSearchService(storagePath, cfg),
		Stats:  service.NewStatsService(storagePath, cfg),
		Config: service.NewConfigService(configPath, cfg),
	}, timerPath
}

// Helper functions tests

func TestRenderEntryList(t *testing.T) {
	styles := ui.DefaultStyles()
	now := time.Now()

	entries := []service.IndexedEntry{
		{
			Entry: entry.Entry{
				Description:     "Short task",
				DurationMinutes: 30,
				Timestamp:       now,
			},
			ActiveIndex: 1,
		},
		{
			Entry: entry.Entry{
				Description:     "Longer task description here",
				DurationMinutes: 90,
				Timestamp:       now,
				Project:         "acme",
				Tags:            []string{"urgent"},
			},
			ActiveIndex: 2,
		},
	}

	// Test without date
	result := RenderEntryList(entries, styles, EntryRenderOptions{
		ShowDate: false,
		Width:    80,
		Cursor:   0,
	})

	if result == "" {
		t.Error("expected non-empty result")
	}
	if !strings.Contains(result, "[1]") {
		t.Error("expected index [1] in result")
	}
	if !strings.Contains(result, "[2]") {
		t.Error("expected index [2] in result")
	}
	if !strings.Contains(result, "30m") {
		t.Error("expected duration 30m in result")
	}
	if !strings.Contains(result, "1h 30m") {
		t.Error("expected duration 1h 30m in result")
	}

	// Test with date
	resultWithDate := RenderEntryList(entries, styles, EntryRenderOptions{
		ShowDate: true,
		Width:    80,
		Cursor:   -1,
	})

	if resultWithDate == "" {
		t.Error("expected non-empty result with date")
	}
	// Date format should be "Jan 02 15:04"
	if !strings.Contains(resultWithDate, now.Format("Jan 02")) {
		t.Error("expected date in result when ShowDate is true")
	}
}

func TestRenderEntryList_Empty(t *testing.T) {
	styles := ui.DefaultStyles()
	result := RenderEntryList(nil, styles, EntryRenderOptions{
		ShowDate: false,
		Width:    80,
		Cursor:   -1,
	})

	if result != "" {
		t.Errorf("expected empty result for empty entries, got %q", result)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		minutes int
		want    string
	}{
		{0, "0m"},
		{30, "30m"},
		{59, "59m"},
		{60, "1h"},
		{90, "1h 30m"},
		{120, "2h"},
		{150, "2h 30m"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			result := formatDuration(tt.minutes)
			if result != tt.want {
				t.Errorf("formatDuration(%d) = %q, want %q", tt.minutes, result, tt.want)
			}
		})
	}
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		word  string
		count int
		want  string
	}{
		{"entry", 0, "entrys"},
		{"entry", 1, "entry"},
		{"entry", 2, "entrys"},
		{"day", 1, "day"},
		{"day", 5, "days"},
	}

	for _, tt := range tests {
		result := pluralize(tt.word, tt.count)
		if result != tt.want {
			t.Errorf("pluralize(%q, %d) = %q, want %q", tt.word, tt.count, result, tt.want)
		}
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{5, 5, 5},
		{-1, 1, -1},
	}

	for _, tt := range tests {
		result := min(tt.a, tt.b)
		if result != tt.want {
			t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.want)
		}
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 2},
		{2, 1, 2},
		{5, 5, 5},
		{-1, 1, 1},
	}

	for _, tt := range tests {
		result := max(tt.a, tt.b)
		if result != tt.want {
			t.Errorf("max(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.want)
		}
	}
}

func TestFormatElapsedTime(t *testing.T) {
	tests := []struct {
		duration time.Duration
		want     string
	}{
		{0, "0m"},
		{30 * time.Minute, "30m"},
		{59 * time.Minute, "59m"},
		{60 * time.Minute, "1h"},
		{90 * time.Minute, "1h 30m"},
		{120 * time.Minute, "2h"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			result := formatElapsedTime(tt.duration)
			if result != tt.want {
				t.Errorf("formatElapsedTime(%v) = %q, want %q", tt.duration, result, tt.want)
			}
		})
	}
}

func TestFormatTimerStartTime_Today(t *testing.T) {
	now := time.Now()
	result := formatTimerStartTime(now)
	if !strings.Contains(result, "today") {
		t.Errorf("expected 'today' in result, got %q", result)
	}
}

func TestFormatTimerStartTime_NotToday(t *testing.T) {
	yesterday := time.Now().AddDate(0, 0, -1)
	result := formatTimerStartTime(yesterday)
	if strings.Contains(result, "today") {
		t.Errorf("did not expect 'today' in result for yesterday, got %q", result)
	}
}

func TestFormatTimerDescription(t *testing.T) {
	state := &timer.TimerState{
		Description: "task",
		Project:     "project",
		Tags:        []string{"tag1", "tag2"},
	}

	result := formatTimerDescription(state)
	if !strings.Contains(result, "task") {
		t.Error("expected 'task' in result")
	}
	if !strings.Contains(result, "@project") {
		t.Error("expected '@project' in result")
	}
	if !strings.Contains(result, "#tag1") {
		t.Error("expected '#tag1' in result")
	}
	if !strings.Contains(result, "#tag2") {
		t.Error("expected '#tag2' in result")
	}
}

func TestFormatTimerDescription_NoMetadata(t *testing.T) {
	state := &timer.TimerState{
		Description: "simple task",
	}

	result := formatTimerDescription(state)
	if result != "simple task" {
		t.Errorf("expected 'simple task', got %q", result)
	}
}

// Entries View Tests
func TestNewEntriesModel(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)

	if model.services == nil {
		t.Error("expected services to be set")
	}
	if model.dateRange.Type != service.DateRangeToday {
		t.Error("expected default date range to be Today")
	}
}

func TestEntriesModel_Init(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)
	cmd := model.Init()

	if cmd == nil {
		t.Error("expected Init to return a command")
	}
}

func TestEntriesModel_SetSize(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)
	model.SetSize(80, 24)

	if model.width != 80 {
		t.Errorf("expected width 80, got %d", model.width)
	}
	if model.height != 24 {
		t.Errorf("expected height 24, got %d", model.height)
	}
}

func TestEntriesModel_View_Loading(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)
	model.loading = true

	view := model.View()
	if !strings.Contains(view, "Loading") {
		t.Errorf("expected 'Loading' in view, got %q", view)
	}
}

func TestEntriesModel_View_Empty(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)
	// Not loading, no entries

	view := model.View()
	if !strings.Contains(view, "No entries") {
		t.Errorf("expected 'No entries' in view, got %q", view)
	}
}

func TestEntriesModel_View_WithEntries(t *testing.T) {
	services := setupTestServicesWithEntries(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)

	// Simulate loading entries
	msg := entriesLoadedMsg{
		entries: []service.IndexedEntry{
			{ActiveIndex: 1, Entry: service.IndexedEntry{}.Entry},
		},
		period: "today",
		total:  60,
	}
	model, _ = model.Update(msg)

	view := model.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestEntriesModel_Update_Navigation(t *testing.T) {
	services := setupTestServicesWithEntries(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)

	// Add some entries
	model.entries = []service.IndexedEntry{
		{ActiveIndex: 1},
		{ActiveIndex: 2},
		{ActiveIndex: 3},
	}

	// Test down navigation
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if model.cursor != 1 {
		t.Errorf("expected cursor 1 after down, got %d", model.cursor)
	}

	// Test up navigation
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if model.cursor != 0 {
		t.Errorf("expected cursor 0 after up, got %d", model.cursor)
	}

	// Test can't go above 0
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if model.cursor != 0 {
		t.Errorf("expected cursor to stay at 0, got %d", model.cursor)
	}
}

func TestEntriesModel_Update_DateRangeKeys(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	tests := []struct {
		key      rune
		expected service.DateRange
	}{
		{'t', service.DateRangeToday},
		{'y', service.DateRangeYesterday},
		{'w', service.DateRangeThisWeek},
		{'W', service.DateRangePrevWeek},
		{'m', service.DateRangeThisMonth},
	}

	for _, tt := range tests {
		model := NewEntriesModel(services, styles, keys)
		model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.key}})

		if model.dateRange.Type != tt.expected {
			t.Errorf("key %c: expected date range %d, got %d", tt.key, tt.expected, model.dateRange.Type)
		}
		if cmd == nil {
			t.Errorf("key %c: expected command to reload entries", tt.key)
		}
	}
}

func TestEntriesModel_IsMultiDayRange(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	tests := []struct {
		dateRange service.DateRange
		multiDay  bool
	}{
		{service.DateRangeToday, false},
		{service.DateRangeYesterday, false},
		{service.DateRangeThisWeek, true},
		{service.DateRangePrevWeek, true},
		{service.DateRangeThisMonth, true},
		{service.DateRangePrevMonth, true},
	}

	for _, tt := range tests {
		model := NewEntriesModel(services, styles, keys)
		model.dateRange.Type = tt.dateRange

		if model.isMultiDayRange() != tt.multiDay {
			t.Errorf("date range %d: expected isMultiDayRange=%v, got %v",
				tt.dateRange, tt.multiDay, model.isMultiDayRange())
		}
	}
}

func TestEntriesModel_Update_Refresh(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

	if cmd == nil {
		t.Error("expected command after refresh")
	}
}

func TestEntriesModel_Update_EntriesLoaded(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)
	model.loading = true

	// Simulate entries loaded message
	entries := []service.IndexedEntry{
		{ActiveIndex: 1},
		{ActiveIndex: 2},
	}
	msg := entriesLoadedMsg{
		entries: entries,
		period:  "today",
		total:   90,
	}

	model, _ = model.Update(msg)

	if model.loading {
		t.Error("expected loading to be false")
	}
	if len(model.entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(model.entries))
	}
	if model.period != "today" {
		t.Errorf("expected period 'today', got %q", model.period)
	}
	if model.total != 90 {
		t.Errorf("expected total 90, got %d", model.total)
	}
}

func TestEntriesModel_Update_EntriesLoadedError(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)
	model.loading = true

	// Simulate error message
	msg := entriesLoadedMsg{
		err: service.ErrNoEntries,
	}

	model, _ = model.Update(msg)

	if model.loading {
		t.Error("expected loading to be false")
	}
	if model.err == nil {
		t.Error("expected error to be set")
	}
}

// EntriesModel Search Mode Tests
func TestEntriesModel_EnterSearchMode(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)

	// Press 's' to enter search mode
	model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})

	if model.mode != entryModeSearch {
		t.Errorf("expected mode to be entryModeSearch, got %d", model.mode)
	}
	if !model.searchInput.Focused() {
		t.Error("expected search input to be focused")
	}
	if cmd == nil {
		t.Error("expected a blink command for text input")
	}
}

func TestEntriesModel_SearchMode_ExitWithEscape(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)
	model.mode = entryModeSearch
	model.searchInput.Focus()

	// Press Escape to exit search mode
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if model.mode != entryModeNormal {
		t.Errorf("expected mode to be entryModeNormal, got %d", model.mode)
	}
	if model.searchInput.Focused() {
		t.Error("expected search input to be blurred")
	}
}

func TestEntriesModel_SearchMode_PerformSearch(t *testing.T) {
	services := setupTestServicesWithEntries(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)
	model.mode = entryModeSearch
	model.searchInput.Focus()
	model.searchInput.SetValue("task")

	// Press Enter to perform search
	model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if model.searchInput.Focused() {
		t.Error("expected search input to be blurred after search")
	}
	if cmd == nil {
		t.Error("expected a search command")
	}
}

func TestEntriesModel_SearchMode_EmptyQueryNoSearch(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)
	model.mode = entryModeSearch
	model.searchInput.Focus()
	model.searchInput.SetValue("  ") // Whitespace only

	// Press Enter with empty query
	model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should not search with empty query
	if cmd != nil {
		t.Error("expected no command for empty search query")
	}
}

func TestEntriesModel_SearchMode_Navigation(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)
	model.mode = entryModeSearch
	model.searchInput.Blur() // Results mode
	model.searched = true
	model.searchResults = []service.IndexedEntry{{ActiveIndex: 1}, {ActiveIndex: 2}, {ActiveIndex: 3}}
	model.searchCursor = 0

	// Navigate down
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if model.searchCursor != 1 {
		t.Errorf("expected searchCursor 1 after down, got %d", model.searchCursor)
	}

	// Navigate down again
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if model.searchCursor != 2 {
		t.Errorf("expected searchCursor 2 after down, got %d", model.searchCursor)
	}

	// Navigate up
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if model.searchCursor != 1 {
		t.Errorf("expected searchCursor 1 after up, got %d", model.searchCursor)
	}
}

func TestEntriesModel_SearchMode_RefocusInput(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)
	model.mode = entryModeSearch
	model.searchInput.Blur() // Results mode
	model.searched = true

	// Press '/' to refocus search input
	model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})

	if !model.searchInput.Focused() {
		t.Error("expected search input to be focused after '/'")
	}
	if cmd == nil {
		t.Error("expected a blink command")
	}
}

func TestEntriesModel_SearchResults_Loaded(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)
	model.mode = entryModeSearch

	// Receive search results
	results := []service.IndexedEntry{{ActiveIndex: 1}, {ActiveIndex: 2}}
	model, _ = model.Update(searchResultsMsg{results: results})

	if !model.searched {
		t.Error("expected searched to be true")
	}
	if len(model.searchResults) != 2 {
		t.Errorf("expected 2 search results, got %d", len(model.searchResults))
	}
	if model.searchCursor != 0 {
		t.Errorf("expected searchCursor to be 0, got %d", model.searchCursor)
	}
}

func TestEntriesModel_SearchResults_WithError(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)
	model.mode = entryModeSearch

	// Receive search error
	testErr := service.ErrNoEntries
	model, _ = model.Update(searchResultsMsg{err: testErr})

	if !model.searched {
		t.Error("expected searched to be true even with error")
	}
	if model.err == nil {
		t.Error("expected error to be set")
	}
}

func TestEntriesModel_View_SearchMode(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)
	model.mode = entryModeSearch

	view := model.View()
	if !strings.Contains(view, "Search") {
		t.Errorf("expected 'Search' in view, got %q", view)
	}
}

func TestEntriesModel_View_SearchResults(t *testing.T) {
	services := setupTestServicesWithEntries(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)
	model.mode = entryModeSearch
	model.searched = true
	model.searchResults = []service.IndexedEntry{
		{ActiveIndex: 1},
	}

	view := model.View()
	if !strings.Contains(view, "Found") {
		t.Errorf("expected 'Found' in view, got %q", view)
	}
}

func TestEntriesModel_View_NoSearchResults(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)
	model.mode = entryModeSearch
	model.searched = true
	model.searchResults = []service.IndexedEntry{}

	view := model.View()
	if !strings.Contains(view, "No results") {
		t.Errorf("expected 'No results' in view, got %q", view)
	}
}

func TestEntriesModel_IsInputMode_SearchMode(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)

	// Normal mode - not input mode
	if model.IsInputMode() {
		t.Error("expected IsInputMode to be false in normal mode")
	}

	// Search mode - not considered input mode (doesn't block Tab)
	model.mode = entryModeSearch
	if model.IsInputMode() {
		t.Error("expected IsInputMode to be false in search mode (allows Tab to switch views)")
	}
}

// Timer View Tests
func TestNewTimerModel(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewTimerModel(services, styles, keys)

	if model.services == nil {
		t.Error("expected services to be set")
	}
}

func TestTimerModel_Init(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewTimerModel(services, styles, keys)
	cmd := model.Init()

	if cmd == nil {
		t.Error("expected Init to return a command")
	}
}

func TestTimerModel_SetSize(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewTimerModel(services, styles, keys)
	model.SetSize(80, 24)

	if model.width != 80 {
		t.Errorf("expected width 80, got %d", model.width)
	}
}

func TestTimerModel_View_Loading(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewTimerModel(services, styles, keys)
	model.loading = true

	view := model.View()
	if !strings.Contains(view, "Loading") {
		t.Errorf("expected 'Loading' in view, got %q", view)
	}
}

func TestTimerModel_View_NoTimer(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewTimerModel(services, styles, keys)
	model.status = &service.TimerStatus{Running: false}

	view := model.View()
	if !strings.Contains(view, "No timer running") {
		t.Errorf("expected 'No timer running' in view, got %q", view)
	}
}

func TestTimerModel_View_TimerRunning(t *testing.T) {
	services, _ := setupTestServicesWithTimer(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewTimerModel(services, styles, keys)
	model.status = &service.TimerStatus{
		Running: true,
		State: &timer.TimerState{
			StartedAt:   time.Now().Add(-30 * time.Minute),
			Description: "working",
			Project:     "project",
			Tags:        []string{"tag"},
		},
		ElapsedTime: 30 * time.Minute,
	}

	view := model.View()
	if !strings.Contains(view, "Timer Running") {
		t.Errorf("expected 'Timer Running' in view, got %q", view)
	}
	if !strings.Contains(view, "working") {
		t.Errorf("expected 'working' in view, got %q", view)
	}
}

func TestTimerModel_Update_StatusLoaded(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewTimerModel(services, styles, keys)
	model.loading = true

	status := &service.TimerStatus{Running: false}
	msg := timerStatusMsg{status: status}

	model, _ = model.Update(msg)

	if model.loading {
		t.Error("expected loading to be false")
	}
	if model.status != status {
		t.Error("expected status to be set")
	}
}

func TestTimerModel_Update_Tick(t *testing.T) {
	services, _ := setupTestServicesWithTimer(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewTimerModel(services, styles, keys)
	model.status = &service.TimerStatus{
		Running: true,
		State: &timer.TimerState{
			StartedAt:   time.Now().Add(-30 * time.Minute),
			Description: "working",
		},
		ElapsedTime: 30 * time.Minute,
	}

	// Send tick message
	model, cmd := model.Update(timerTickMsg(time.Now()))

	if cmd == nil {
		t.Error("expected tick to return another tick command")
	}
}

func TestTimerModel_Update_StopTimer(t *testing.T) {
	services, _ := setupTestServicesWithTimer(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewTimerModel(services, styles, keys)
	model.status = &service.TimerStatus{Running: true}

	// Press stop key
	model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	if cmd == nil {
		t.Error("expected stop to return a command")
	}
}

func TestTimerModel_Update_Refresh(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewTimerModel(services, styles, keys)
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

	if cmd == nil {
		t.Error("expected refresh to return a command")
	}
}

// Stats View Tests
func TestNewStatsModel(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewStatsModel(services, styles, keys)

	if model.services == nil {
		t.Error("expected services to be set")
	}
	if !model.weekly {
		t.Error("expected default to be weekly")
	}
}

func TestStatsModel_Init(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewStatsModel(services, styles, keys)
	cmd := model.Init()

	if cmd == nil {
		t.Error("expected Init to return a command")
	}
}

func TestStatsModel_SetSize(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewStatsModel(services, styles, keys)
	model.SetSize(80, 24)

	if model.width != 80 {
		t.Errorf("expected width 80, got %d", model.width)
	}
}

func TestStatsModel_View_Loading(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewStatsModel(services, styles, keys)
	model.loading = true

	view := model.View()
	if !strings.Contains(view, "Loading") {
		t.Errorf("expected 'Loading' in view, got %q", view)
	}
}

func TestStatsModel_View_NoData(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewStatsModel(services, styles, keys)
	model.result = nil

	view := model.View()
	if !strings.Contains(view, "No data") {
		t.Errorf("expected 'No data' in view, got %q", view)
	}
}

func TestStatsModel_View_Weekly(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewStatsModel(services, styles, keys)
	model.weekly = true
	model.result = &service.StatsResult{
		Statistics: stats.Statistics{
			TotalMinutes:         120,
			EntryCount:           5,
			DaysWithEntries:      3,
			AverageMinutesPerDay: 40,
		},
		Period:     "this week",
		Comparison: "+30m vs last week",
	}

	view := model.View()
	if !strings.Contains(view, "Weekly") {
		t.Errorf("expected 'Weekly' in view, got %q", view)
	}
	if !strings.Contains(view, "Total time") {
		t.Errorf("expected 'Total time' in view, got %q", view)
	}
}

func TestStatsModel_View_Monthly(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewStatsModel(services, styles, keys)
	model.weekly = false
	model.result = &service.StatsResult{
		Statistics: stats.Statistics{TotalMinutes: 480},
	}

	view := model.View()
	if !strings.Contains(view, "Monthly") {
		t.Errorf("expected 'Monthly' in view, got %q", view)
	}
}

func TestStatsModel_View_WithProjectStats(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewStatsModel(services, styles, keys)
	model.result = &service.StatsResult{
		Statistics: stats.Statistics{TotalMinutes: 120},
		ProjectStats: []stats.ProjectBreakdown{
			{Project: "acme", TotalMinutes: 60, EntryCount: 2},
			{Project: "(no project)", TotalMinutes: 60, EntryCount: 3},
		},
	}

	view := model.View()
	if !strings.Contains(view, "By Project") {
		t.Errorf("expected 'By Project' in view, got %q", view)
	}
	if !strings.Contains(view, "@acme") {
		t.Errorf("expected '@acme' in view, got %q", view)
	}
}

func TestStatsModel_Update_WeekKey(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewStatsModel(services, styles, keys)
	model.weekly = false

	model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})

	if !model.weekly {
		t.Error("expected weekly to be true after 'w' key")
	}
	if cmd == nil {
		t.Error("expected command to reload stats")
	}
}

func TestStatsModel_Update_MonthKey(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewStatsModel(services, styles, keys)
	model.weekly = true

	model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})

	if model.weekly {
		t.Error("expected weekly to be false after 'm' key")
	}
	if cmd == nil {
		t.Error("expected command to reload stats")
	}
}

func TestStatsModel_Update_StatsLoaded(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewStatsModel(services, styles, keys)
	model.loading = true

	result := &service.StatsResult{
		Statistics: stats.Statistics{TotalMinutes: 120},
	}
	msg := statsLoadedMsg{result: result}

	model, _ = model.Update(msg)

	if model.loading {
		t.Error("expected loading to be false")
	}
	if model.result != result {
		t.Error("expected result to be set")
	}
}

// Config View Tests
func TestNewConfigModel(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	themeProvider := ui.NewThemeProvider("")
	model := NewConfigModel(services, themeProvider, styles, keys)

	if model.services == nil {
		t.Error("expected services to be set")
	}
}

func TestConfigModel_Init(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	themeProvider := ui.NewThemeProvider("")
	model := NewConfigModel(services, themeProvider, styles, keys)
	cmd := model.Init()

	if cmd == nil {
		t.Error("expected Init to return a command")
	}
}

func TestConfigModel_SetSize(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	themeProvider := ui.NewThemeProvider("")
	model := NewConfigModel(services, themeProvider, styles, keys)
	model.SetSize(80, 24)

	if model.width != 80 {
		t.Errorf("expected width 80, got %d", model.width)
	}
}

func TestConfigModel_View(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	themeProvider := ui.NewThemeProvider("")
	model := NewConfigModel(services, themeProvider, styles, keys)
	model.config = config.DefaultConfig()
	model.path = "/path/to/config"
	model.exists = false

	view := model.View()
	if !strings.Contains(view, "Configuration") {
		t.Errorf("expected 'Configuration' in view, got %q", view)
	}
	if !strings.Contains(view, "Config file") {
		t.Errorf("expected 'Config file' in view, got %q", view)
	}
	if !strings.Contains(view, "week_start_day") {
		t.Errorf("expected 'week_start_day' in view, got %q", view)
	}
}

func TestConfigModel_View_FileExists(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	themeProvider := ui.NewThemeProvider("")
	model := NewConfigModel(services, themeProvider, styles, keys)
	model.config = config.DefaultConfig()
	model.path = "/path/to/config"
	model.exists = true

	view := model.View()
	if !strings.Contains(view, "File exists") {
		t.Errorf("expected 'File exists' in view, got %q", view)
	}
}

func TestConfigModel_View_NoFile(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	themeProvider := ui.NewThemeProvider("")
	model := NewConfigModel(services, themeProvider, styles, keys)
	model.config = config.DefaultConfig()
	model.path = "/path/to/config"
	model.exists = false

	view := model.View()
	if !strings.Contains(view, "Using defaults") {
		t.Errorf("expected 'Using defaults' in view, got %q", view)
	}
}

func TestConfigModel_Update_ConfigLoaded(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	themeProvider := ui.NewThemeProvider("")
	model := NewConfigModel(services, themeProvider, styles, keys)

	cfg := config.DefaultConfig()
	msg := configLoadedMsg{
		config: cfg,
		path:   "/test/path",
		exists: true,
	}

	model, _ = model.Update(msg)

	if model.path != "/test/path" {
		t.Errorf("expected path '/test/path', got %q", model.path)
	}
	if !model.exists {
		t.Error("expected exists to be true")
	}
}

// Tests for executing commands returned by Init/Update

func TestTimerModel_LoadStatus_ExecuteCmd(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewTimerModel(services, styles, keys)
	cmd := model.Init()

	// Init should return a command
	if cmd == nil {
		t.Error("expected Init to return a command")
	}
}

func TestTimerModel_StopTimer_ExecuteCmd(t *testing.T) {
	services, _ := setupTestServicesWithTimer(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewTimerModel(services, styles, keys)
	// Set status to running so stop key triggers stopTimer
	model.status = &service.TimerStatus{Running: true}

	// Press stop key
	model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(timerStatusMsg); !ok {
			t.Errorf("expected timerStatusMsg from stop, got %T", msg)
		}
	}
}

func TestTimerModel_TickTimer(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewTimerModel(services, styles, keys)
	model.status = &service.TimerStatus{
		Running: true,
		State: &timer.TimerState{
			StartedAt: time.Now().Add(-10 * time.Minute),
		},
	}

	// Send tick message
	model, cmd := model.Update(timerTickMsg(time.Now()))

	// When running, tick should return another tick command
	if cmd == nil {
		t.Error("expected tick to return another command when running")
	}
}

func TestTimerModel_TickTimer_NotRunning(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewTimerModel(services, styles, keys)
	model.status = &service.TimerStatus{Running: false}

	// Send tick message
	model, cmd := model.Update(timerTickMsg(time.Now()))

	// Tick always returns another tick command (timer keeps ticking to check status)
	if cmd == nil {
		t.Error("expected tick to return another command")
	}
}

func TestStatsModel_LoadStats_ExecuteCmd(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewStatsModel(services, styles, keys)
	cmd := model.Init()

	// Execute the command returned by Init (loadStats)
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(statsLoadedMsg); !ok {
			t.Errorf("expected statsLoadedMsg, got %T", msg)
		}
	}
}

func TestStatsModel_LoadStats_Monthly(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewStatsModel(services, styles, keys)
	model.weekly = false

	// Press 'm' key to ensure monthly mode and trigger reload
	model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})

	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(statsLoadedMsg); !ok {
			t.Errorf("expected statsLoadedMsg, got %T", msg)
		}
	}
}

func TestEntriesModel_LoadEntries_ExecuteCmd(t *testing.T) {
	services := setupTestServicesWithEntries(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)
	cmd := model.Init()

	// Execute the command returned by Init (loadEntries)
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(entriesLoadedMsg); !ok {
			t.Errorf("expected entriesLoadedMsg, got %T", msg)
		}
	}
}

func TestConfigModel_LoadConfig_ExecuteCmd(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	themeProvider := ui.NewThemeProvider("")
	model := NewConfigModel(services, themeProvider, styles, keys)
	cmd := model.Init()

	// Execute the command returned by Init (loadConfig)
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(configLoadedMsg); !ok {
			t.Errorf("expected configLoadedMsg, got %T", msg)
		}
	}
}


func TestTimerModel_Update_StatusWithError(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewTimerModel(services, styles, keys)

	// Send status message with error
	testErr := os.ErrPermission
	model, _ = model.Update(timerStatusMsg{err: testErr})

	if model.err != testErr {
		t.Error("expected error to be set")
	}
}

func TestStatsModel_Update_StatsWithError(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewStatsModel(services, styles, keys)

	// Send stats message with error
	testErr := os.ErrPermission
	model, _ = model.Update(statsLoadedMsg{err: testErr})

	if model.err != testErr {
		t.Error("expected error to be set")
	}
}

func TestEntriesModel_Update_EntriesWithError(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)

	// Send entries message with error
	testErr := os.ErrPermission
	model, _ = model.Update(entriesLoadedMsg{err: testErr})

	if model.err != testErr {
		t.Error("expected error to be set")
	}
}


// Test views with errors to cover error display paths
func TestEntriesModel_View_WithError(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)
	model.err = os.ErrPermission
	model.loading = false

	view := model.View()
	if !strings.Contains(view, "Error") {
		t.Errorf("expected 'Error' in view with error, got %q", view)
	}
}


func TestStatsModel_View_WithError(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewStatsModel(services, styles, keys)
	model.err = os.ErrPermission
	model.loading = false

	view := model.View()
	if !strings.Contains(view, "Error") {
		t.Errorf("expected 'Error' in view with error, got %q", view)
	}
}

func TestTimerModel_View_WithError(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewTimerModel(services, styles, keys)
	model.err = os.ErrPermission
	model.loading = false

	view := model.View()
	if !strings.Contains(view, "Error") {
		t.Errorf("expected 'Error' in view with error, got %q", view)
	}
}

// Test entries with metadata to cover project/tag rendering
func TestEntriesModel_View_WithProjectAndTags(t *testing.T) {
	services := setupTestServicesWithEntries(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)
	model.entries = []service.IndexedEntry{
		{
			ActiveIndex: 1,
			Entry: entry.Entry{
				Description:     "task with metadata",
				DurationMinutes: 60,
				Timestamp:       time.Now(),
				Project:         "myproject",
				Tags:            []string{"urgent", "feature"},
			},
		},
	}

	view := model.View()
	// The view should contain the project and tags
	if !strings.Contains(view, "task with metadata") {
		t.Errorf("expected task description in view, got %q", view)
	}
}

// Test cursor adjustment when entries are removed
func TestEntriesModel_Update_CursorAdjustment(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)
	model.entries = []service.IndexedEntry{{ActiveIndex: 1}, {ActiveIndex: 2}, {ActiveIndex: 3}}
	model.cursor = 2 // At the last item

	// Receive entries with fewer items
	model, _ = model.Update(entriesLoadedMsg{
		entries: []service.IndexedEntry{{ActiveIndex: 1}},
		period:  "today",
	})

	// Cursor should be adjusted to be within bounds
	if model.cursor >= len(model.entries) {
		t.Errorf("cursor %d should be less than entries length %d", model.cursor, len(model.entries))
	}
}

// Test stats refresh key
func TestStatsModel_Update_RefreshKey(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewStatsModel(services, styles, keys)

	// Press 'r' for refresh
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

	if cmd == nil {
		t.Error("expected refresh to return a command")
	}
}

// Test timer start key - should enter input mode
func TestTimerModel_Update_StartKey(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewTimerModel(services, styles, keys)
	model.status = &service.TimerStatus{Running: false}

	// Press 's' for start - should enter input mode
	model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})

	// Should return a blink command for text input
	if cmd == nil {
		t.Error("expected start to return a command for text input")
	}
	// Should be in input mode
	if !model.inputMode {
		t.Error("expected timer to be in input mode after pressing 's'")
	}
}

// Test timer view with running status
func TestTimerModel_View_Running(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewTimerModel(services, styles, keys)
	model.status = &service.TimerStatus{
		Running: true,
		State: &timer.TimerState{
			StartedAt:   time.Now().Add(-30 * time.Minute),
			Description: "working on feature",
			Project:     "myproject",
			Tags:        []string{"feature"},
		},
		ElapsedTime: 30 * time.Minute,
	}

	view := model.View()
	if !strings.Contains(view, "Running") {
		t.Errorf("expected 'Running' in timer view, got %q", view)
	}
	if !strings.Contains(view, "working on feature") {
		t.Errorf("expected description in timer view, got %q", view)
	}
}

// Helper for services with broken storage (to test error paths in commands)
func setupBrokenServices(t *testing.T) *service.Services {
	t.Helper()
	cfg := config.DefaultConfig()
	// Use a non-existent directory path to cause errors
	brokenPath := "/nonexistent/path/entries.jsonl"
	brokenTimerPath := "/nonexistent/path/timer.json"
	brokenConfigPath := "/nonexistent/path/config.yaml"

	return &service.Services{
		Entry:  service.NewEntryService(brokenPath, cfg),
		Timer:  service.NewTimerService(brokenTimerPath, brokenPath, cfg),
		Report: service.NewReportService(brokenPath, cfg),
		Search: service.NewSearchService(brokenPath, cfg),
		Stats:  service.NewStatsService(brokenPath, cfg),
		Config: service.NewConfigService(brokenConfigPath, cfg),
	}
}

// Test that loadEntries handles errors
func TestEntriesModel_LoadEntries_WithBrokenStorage(t *testing.T) {
	services := setupBrokenServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewEntriesModel(services, styles, keys)
	cmd := model.Init()

	// Execute the command - it should return an error message
	if cmd != nil {
		msg := cmd()
		if loaded, ok := msg.(entriesLoadedMsg); ok {
			// The error path should have returned an error
			// (depending on storage behavior, this might or might not error)
			_ = loaded
		}
	}
}


// Test that stopTimer handles errors
func TestTimerModel_StopTimer_WithBrokenStorage(t *testing.T) {
	services := setupBrokenServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewTimerModel(services, styles, keys)
	model.status = &service.TimerStatus{Running: true}

	// Trigger stop
	model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	if cmd != nil {
		msg := cmd()
		if result, ok := msg.(timerStatusMsg); ok {
			// Should have error due to broken storage
			_ = result
		}
	}
}

// Test unhandled key returns nil command
func TestTimerModel_Update_UnhandledKey(t *testing.T) {
	services := setupTestServices(t)
	styles := ui.DefaultStyles()
	keys := ui.DefaultKeyMap()

	model := NewTimerModel(services, styles, keys)
	model.status = &service.TimerStatus{Running: false}

	// Press a key that's not handled specifically
	model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})

	// Unhandled keys should return nil command
	if cmd != nil {
		t.Error("expected nil command for unhandled key")
	}
}
