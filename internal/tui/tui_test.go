package tui

import (
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xolan/did/internal/config"
	"github.com/xolan/did/internal/service"
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

func TestNew(t *testing.T) {
	services := setupTestServices(t)
	model := New(services)

	if model.activeTab != TabEntries {
		t.Errorf("expected initial tab to be Entries, got %d", model.activeTab)
	}
	if model.services == nil {
		t.Error("expected services to be set")
	}
	if model.showHelp {
		t.Error("expected showHelp to be false initially")
	}
}

func TestInit(t *testing.T) {
	services := setupTestServices(t)
	model := New(services)

	cmd := model.Init()
	if cmd == nil {
		t.Error("expected Init to return a command")
	}
}

func TestUpdate_WindowSizeMsg(t *testing.T) {
	services := setupTestServices(t)
	model := New(services)

	// Send window size message
	newModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	m := newModel.(Model)

	if m.width != 100 {
		t.Errorf("expected width 100, got %d", m.width)
	}
	if m.height != 50 {
		t.Errorf("expected height 50, got %d", m.height)
	}
}

func TestUpdate_QuitKey(t *testing.T) {
	services := setupTestServices(t)
	model := New(services)

	// Send quit key
	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	_ = newModel

	// Quit should return a tea.Quit command
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestUpdate_HelpKey(t *testing.T) {
	services := setupTestServices(t)
	model := New(services)

	// Send help key
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m := newModel.(Model)

	if !m.showHelp {
		t.Error("expected showHelp to be true after pressing ?")
	}

	// Toggle off
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = newModel.(Model)

	if m.showHelp {
		t.Error("expected showHelp to be false after pressing ? again")
	}
}

func TestUpdate_TabNavigation(t *testing.T) {
	services := setupTestServices(t)
	model := New(services)

	// Initial tab should be Entries
	if model.activeTab != TabEntries {
		t.Errorf("expected initial tab TabEntries, got %d", model.activeTab)
	}

	// Press tab to go to next tab
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	m := newModel.(Model)

	if m.activeTab != TabTimer {
		t.Errorf("expected TabTimer after pressing tab, got %d", m.activeTab)
	}
}

func TestUpdate_DirectTabKeys(t *testing.T) {
	services := setupTestServices(t)
	model := New(services)

	tests := []struct {
		key      rune
		expected Tab
	}{
		{'1', TabEntries},
		{'2', TabTimer},
		{'3', TabStats},
		{'4', TabConfig},
	}

	for _, tt := range tests {
		newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.key}})
		m := newModel.(Model)

		if m.activeTab != tt.expected {
			t.Errorf("pressing %c: expected tab %d, got %d", tt.key, tt.expected, m.activeTab)
		}
	}
}

func TestUpdate_PrevTab(t *testing.T) {
	services := setupTestServices(t)
	model := New(services)

	// Go to second tab first
	model.activeTab = TabTimer

	// Press shift+tab to go to previous tab
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m := newModel.(Model)

	if m.activeTab != TabEntries {
		t.Errorf("expected TabEntries after shift+tab, got %d", m.activeTab)
	}
}

func TestUpdate_PrevTab_Wraparound(t *testing.T) {
	services := setupTestServices(t)
	model := New(services)

	// Start at first tab
	model.activeTab = TabEntries

	// Press shift+tab should wrap to last tab
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m := newModel.(Model)

	if m.activeTab != TabConfig {
		t.Errorf("expected TabConfig (wraparound) after shift+tab from TabEntries, got %d", m.activeTab)
	}
}

func TestUpdate_NextTab_Wraparound(t *testing.T) {
	services := setupTestServices(t)
	model := New(services)

	// Start at last tab
	model.activeTab = TabConfig

	// Press tab should wrap to first tab
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	m := newModel.(Model)

	if m.activeTab != TabEntries {
		t.Errorf("expected TabEntries (wraparound) after tab from TabConfig, got %d", m.activeTab)
	}
}

func TestView_Loading(t *testing.T) {
	services := setupTestServices(t)
	model := New(services)

	// Before window size is set, width is 0
	view := model.View()

	if !strings.Contains(view, "Loading") {
		t.Errorf("expected 'Loading...' when width is 0, got %q", view)
	}
}

func TestView_WithSize(t *testing.T) {
	services := setupTestServices(t)
	model := New(services)

	// Set window size
	newModel, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m := newModel.(Model)

	view := m.View()

	// Should contain tab names
	if !strings.Contains(view, "Entries") {
		t.Error("expected 'Entries' tab in view")
	}

	// Should contain status bar help text
	if !strings.Contains(view, "quit") {
		t.Error("expected 'quit' in status bar")
	}
}

func TestView_AllTabs(t *testing.T) {
	services := setupTestServices(t)
	model := New(services)

	// Set window size
	newModel, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m := newModel.(Model)

	// Test view for each tab
	tabs := []Tab{TabEntries, TabTimer, TabStats, TabConfig}
	for _, tab := range tabs {
		m.activeTab = tab
		view := m.View()

		if view == "" {
			t.Errorf("expected non-empty view for tab %d", tab)
		}
	}
}

func TestRenderTabs(t *testing.T) {
	services := setupTestServices(t)
	model := New(services)

	tabs := model.renderTabs()

	// Should contain all tab names
	for _, name := range tabNames {
		if !strings.Contains(tabs, name) {
			t.Errorf("expected tab name %s in rendered tabs", name)
		}
	}
}

func TestRenderStatusBar(t *testing.T) {
	services := setupTestServices(t)
	model := New(services)
	model.width = 80

	statusBar := model.renderStatusBar()

	// Should contain common keys
	if !strings.Contains(statusBar, "1-4") {
		t.Error("expected '1-4' in status bar")
	}
	if !strings.Contains(statusBar, "quit") {
		t.Error("expected 'quit' in status bar")
	}
	if !strings.Contains(statusBar, "?") {
		t.Error("expected '?' in status bar")
	}
}

func TestRenderStatusBar_EntryTab(t *testing.T) {
	services := setupTestServices(t)
	model := New(services)
	model.width = 80
	model.activeTab = TabEntries

	statusBar := model.renderStatusBar()

	// Entry tab specific keys
	if !strings.Contains(statusBar, "new") {
		t.Error("expected 'new' in status bar for entries tab")
	}
	if !strings.Contains(statusBar, "search") {
		t.Error("expected 'search' in status bar for entries tab")
	}
	if !strings.Contains(statusBar, "filter") {
		t.Error("expected 'filter' in status bar for entries tab")
	}
}

func TestRenderStatusBar_TimerTab(t *testing.T) {
	services := setupTestServices(t)
	model := New(services)
	model.width = 80
	model.activeTab = TabTimer

	statusBar := model.renderStatusBar()

	// Timer tab specific keys
	if !strings.Contains(statusBar, "start") {
		t.Error("expected 'start' in status bar for timer tab")
	}
	if !strings.Contains(statusBar, "stop") {
		t.Error("expected 'stop' in status bar for timer tab")
	}
}

func TestRenderStatusBar_StatsTab(t *testing.T) {
	services := setupTestServices(t)
	model := New(services)
	model.width = 80
	model.activeTab = TabStats

	statusBar := model.renderStatusBar()

	// Stats tab specific keys
	if !strings.Contains(statusBar, "week") {
		t.Error("expected 'week' in status bar for stats tab")
	}
	if !strings.Contains(statusBar, "month") {
		t.Error("expected 'month' in status bar for stats tab")
	}
}

func TestRenderKeyHelp(t *testing.T) {
	services := setupTestServices(t)
	model := New(services)

	help := model.renderKeyHelp("q", "quit")

	if !strings.Contains(help, "q") {
		t.Error("expected key 'q' in key help")
	}
	if !strings.Contains(help, "quit") {
		t.Error("expected description 'quit' in key help")
	}
}

func TestInitCurrentView(t *testing.T) {
	services := setupTestServices(t)
	model := New(services)

	// Test init for each tab
	tabs := []Tab{TabEntries, TabTimer, TabStats, TabConfig}
	for _, tab := range tabs {
		model.activeTab = tab
		cmd := model.initCurrentView()
		// Some views may return nil, others return a command
		_ = cmd
	}
}

func TestInitCurrentView_InvalidTab(t *testing.T) {
	services := setupTestServices(t)
	model := New(services)

	// Set an invalid tab value
	model.activeTab = Tab(999)
	cmd := model.initCurrentView()

	// Should return nil for invalid tab
	if cmd != nil {
		t.Error("expected nil command for invalid tab")
	}
}

func TestTabNames(t *testing.T) {
	expectedNames := []string{"Entries", "Timer", "Stats", "Config"}

	if len(tabNames) != len(expectedNames) {
		t.Errorf("expected %d tab names, got %d", len(expectedNames), len(tabNames))
	}

	for i, name := range expectedNames {
		if tabNames[i] != name {
			t.Errorf("expected tab name %d to be %s, got %s", i, name, tabNames[i])
		}
	}
}

func TestTabConstants(t *testing.T) {
	// Verify tab constants are sequential
	if TabEntries != 0 {
		t.Error("TabEntries should be 0")
	}
	if TabTimer != 1 {
		t.Error("TabTimer should be 1")
	}
	if TabStats != 2 {
		t.Error("TabStats should be 2")
	}
	if TabConfig != 3 {
		t.Error("TabConfig should be 3")
	}
}

func TestUpdate_PassesMessagesToViews(t *testing.T) {
	services := setupTestServices(t)
	model := New(services)

	// Set size first
	newModel, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m := newModel.(Model)

	// Send a key that would be handled by the view
	// This exercises the view update code path
	tabs := []Tab{TabEntries, TabTimer, TabStats, TabConfig}
	for _, tab := range tabs {
		m.activeTab = tab
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}) // Down key
		m = newModel.(Model)
	}
}

func TestUpdate_ModalInputBlocksAllKeys(t *testing.T) {
	services := setupTestServices(t)
	model := New(services)

	// Set size first
	newModel, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m := newModel.(Model)

	// Go to entries tab and enter add mode by pressing 'n'
	m.activeTab = TabEntries
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = newModel.(Model)

	// Pressing '2' should NOT switch to Timer tab because we're in modal add entry mode
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m = newModel.(Model)

	if m.activeTab != TabEntries {
		t.Errorf("expected to stay on TabEntries when in modal input mode, got %d", m.activeTab)
	}

	// Tab should also NOT switch views in modal input mode
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = newModel.(Model)

	if m.activeTab != TabEntries {
		t.Errorf("expected Tab to NOT switch views in modal input mode, got %d", m.activeTab)
	}
}


func TestUpdate_EntriesViewDateRangeKey(t *testing.T) {
	services := setupTestServices(t)
	model := New(services)

	// Set size first
	newModel, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m := newModel.(Model)

	// Ensure we're on entries tab
	m.activeTab = TabEntries

	// Initial date range should be Today
	if m.entriesView.GetDateRangeType() != service.DateRangeToday {
		t.Fatalf("expected initial date range to be Today, got %d", m.entriesView.GetDateRangeType())
	}

	// Press 'y' for yesterday
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m = newModel.(Model)

	// Date range should now be Yesterday
	if m.entriesView.GetDateRangeType() != service.DateRangeYesterday {
		t.Errorf("expected date range to be Yesterday after 'y' key, got %d", m.entriesView.GetDateRangeType())
	}

	// Should return a command to reload entries
	if cmd == nil {
		t.Error("expected command to reload entries after date range change")
	}

	// Press 'w' for this week
	newModel, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	m = newModel.(Model)

	if m.entriesView.GetDateRangeType() != service.DateRangeThisWeek {
		t.Errorf("expected date range to be ThisWeek after 'w' key, got %d", m.entriesView.GetDateRangeType())
	}

	if cmd == nil {
		t.Error("expected command to reload entries after date range change")
	}
}
