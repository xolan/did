// Package tui provides the Terminal User Interface for the did application.
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xolan/did/internal/service"
	"github.com/xolan/did/internal/tui/ui"
	"github.com/xolan/did/internal/tui/views"
)

// Tab represents a view tab
type Tab int

const (
	TabEntries Tab = iota
	TabTimer
	TabStats
	TabConfig
)

var tabNames = []string{"Entries", "Timer", "Stats", "Config"}

// Model is the root TUI model
type Model struct {
	// Services
	services *service.Services

	// UI state
	activeTab Tab
	width     int
	height    int
	showHelp  bool

	// View models
	entriesView views.EntriesModel
	timerView   views.TimerModel
	statsView   views.StatsModel
	configView  views.ConfigModel

	// Theme and styles
	themeProvider *ui.ThemeProvider
	styles        ui.Styles
	keys          ui.KeyMap
}

// New creates a new TUI model
func New(services *service.Services) Model {
	// Initialize theme from config
	themeName := services.Config.Get().Theme
	themeProvider := ui.NewThemeProvider(themeName)
	styles := themeProvider.Styles()
	keys := ui.DefaultKeyMap()

	return Model{
		services:      services,
		activeTab:     TabEntries,
		themeProvider: themeProvider,
		styles:        styles,
		keys:          keys,
		entriesView:   views.NewEntriesModel(services, styles, keys),
		timerView:     views.NewTimerModel(services, styles, keys),
		statsView:     views.NewStatsModel(services, styles, keys),
		configView:    views.NewConfigModel(services, themeProvider, styles, keys),
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.entriesView.Init(),
		m.timerView.Init(),
	)
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Check input modes:
		// - modalInput: blocks ALL global keys (add/edit entry, timer start)
		// - capturingKeys: blocks character keys but allows Tab (search input)
		modalInput := m.isModalInputMode()
		capturingKeys := m.isCapturingKeys()

		// Handle global keys first
		switch {
		case key.Matches(msg, m.keys.Quit) && !capturingKeys:
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help) && !capturingKeys:
			m.showHelp = !m.showHelp
			return m, nil

		case key.Matches(msg, m.keys.NextTab) && !modalInput:
			m.activeTab = Tab((int(m.activeTab) + 1) % len(tabNames))
			return m, m.initCurrentView()

		case key.Matches(msg, m.keys.PrevTab) && !modalInput:
			m.activeTab = Tab((int(m.activeTab) - 1 + len(tabNames)) % len(tabNames))
			return m, m.initCurrentView()

		case key.Matches(msg, m.keys.Tab1) && !capturingKeys:
			m.activeTab = TabEntries
			return m, m.initCurrentView()

		case key.Matches(msg, m.keys.Tab2) && !capturingKeys:
			m.activeTab = TabTimer
			return m, m.initCurrentView()

		case key.Matches(msg, m.keys.Tab3) && !capturingKeys:
			m.activeTab = TabStats
			return m, m.initCurrentView()

		case key.Matches(msg, m.keys.Tab4) && !capturingKeys:
			m.activeTab = TabConfig
			return m, m.initCurrentView()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update view dimensions
		contentHeight := m.height - 4 // Account for tabs and status bar
		m.entriesView.SetSize(m.width, contentHeight)
		m.timerView.SetSize(m.width, contentHeight)
		m.statsView.SetSize(m.width, contentHeight)
		m.configView.SetSize(m.width, contentHeight)
		return m, nil

	case ui.ThemeChangeRequestMsg:
		// Handle theme change request
		m.themeProvider.SetTheme(msg.ThemeName)
		newTheme := m.themeProvider.CurrentName()

		// Update styles
		m.styles = m.themeProvider.Styles()

		// Broadcast theme change to all views
		themeMsg := ui.ThemeChangedMsg{
			ThemeName: newTheme,
			Styles:    m.styles,
		}
		m.entriesView, _ = m.entriesView.Update(themeMsg)
		m.timerView, _ = m.timerView.Update(themeMsg)
		m.statsView, _ = m.statsView.Update(themeMsg)
		m.configView, _ = m.configView.Update(themeMsg)

		// Save theme to config
		return m, m.saveThemeConfig(newTheme)
	}

	// Update the active view
	switch m.activeTab {
	case TabEntries:
		m.entriesView, cmd = m.entriesView.Update(msg)
	case TabTimer:
		m.timerView, cmd = m.timerView.Update(msg)
	case TabStats:
		m.statsView, cmd = m.statsView.Update(msg)
	case TabConfig:
		m.configView, cmd = m.configView.Update(msg)
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var b strings.Builder

	// Render tabs
	b.WriteString(m.renderTabs())
	b.WriteString("\n")

	// Render active view
	switch m.activeTab {
	case TabEntries:
		b.WriteString(m.entriesView.View())
	case TabTimer:
		b.WriteString(m.timerView.View())
	case TabStats:
		b.WriteString(m.statsView.View())
	case TabConfig:
		b.WriteString(m.configView.View())
	}

	// Render status bar
	b.WriteString("\n")
	b.WriteString(m.renderStatusBar())

	// Help overlay
	if m.showHelp {
		return m.renderHelpOverlay(b.String())
	}

	return m.styles.App.Render(b.String())
}

// renderTabs renders the tab bar
func (m Model) renderTabs() string {
	var tabs []string
	for i, name := range tabNames {
		if Tab(i) == m.activeTab {
			tabs = append(tabs, m.styles.TabActive.Render(name))
		} else {
			tabs = append(tabs, m.styles.TabInactive.Render(name))
		}
	}
	return m.styles.TabBar.Render(lipgloss.JoinHorizontal(lipgloss.Top, tabs...))
}

// renderStatusBar renders the status bar at the bottom
func (m Model) renderStatusBar() string {
	var parts []string

	// Check if in modal input mode for context-specific hints
	if m.isModalInputMode() {
		parts = append(parts, m.renderKeyHelp("Tab", "switch field"))
		parts = append(parts, m.renderKeyHelp("Enter", "save"))
		parts = append(parts, m.renderKeyHelp("Esc", "cancel"))
	} else {
		// View-specific keys
		switch m.activeTab {
		case TabEntries:
			parts = append(parts, m.renderKeyHelp("n", "new"))
			parts = append(parts, m.renderKeyHelp("e", "edit"))
			parts = append(parts, m.renderKeyHelp("d", "delete"))
			parts = append(parts, m.renderKeyHelp("s", "search"))
			parts = append(parts, m.renderKeyHelp("t/y/w", "filter"))
		case TabTimer:
			parts = append(parts, m.renderKeyHelp("s", "start"))
			parts = append(parts, m.renderKeyHelp("x", "stop"))
			parts = append(parts, m.renderKeyHelp("r", "refresh"))
		case TabStats:
			parts = append(parts, m.renderKeyHelp("w", "week"))
			parts = append(parts, m.renderKeyHelp("m", "month"))
		case TabConfig:
			parts = append(parts, m.renderKeyHelp("t", "themes"))
		}

		// Global keys
		parts = append(parts, m.renderKeyHelp("1-4", "views"))
		parts = append(parts, m.renderKeyHelp("?", "help"))
		parts = append(parts, m.renderKeyHelp("q", "quit"))
	}

	content := strings.Join(parts, "  ")

	// Fill to width
	padding := m.width - lipgloss.Width(content)
	if padding > 0 {
		content += strings.Repeat(" ", padding)
	}

	return m.styles.StatusBar.Render(content)
}

// renderKeyHelp renders a single key help item
func (m Model) renderKeyHelp(key, desc string) string {
	return fmt.Sprintf("%s %s",
		m.styles.StatusKey.Render(key),
		m.styles.StatusHelp.Render(desc))
}

// isModalInputMode checks if the current view is in a modal input mode
// where the user should not be able to switch views (add/edit entry, timer start)
func (m Model) isModalInputMode() bool {
	switch m.activeTab {
	case TabEntries:
		return m.entriesView.IsInputMode()
	case TabTimer:
		return m.timerView.IsInputMode()
	}
	return false
}

// isCapturingKeys checks if the current view is capturing keyboard input
func (m Model) isCapturingKeys() bool {
	switch m.activeTab {
	case TabEntries:
		return m.entriesView.IsInputMode()
	case TabTimer:
		return m.timerView.IsInputMode()
	}
	return false
}

// initCurrentView initializes the current view when switching tabs
func (m Model) initCurrentView() tea.Cmd {
	switch m.activeTab {
	case TabEntries:
		return m.entriesView.Init()
	case TabTimer:
		return m.timerView.Init()
	case TabStats:
		return m.statsView.Init()
	case TabConfig:
		return m.configView.Init()
	}
	return nil
}

// saveThemeConfig saves the theme to the config file
func (m Model) saveThemeConfig(themeName string) tea.Cmd {
	return func() tea.Msg {
		cfg := m.services.Config.Get()
		cfg.Theme = themeName
		_ = m.services.Config.Update(cfg)
		return nil
	}
}

// GetThemeProvider returns the theme provider for use by views
func (m Model) GetThemeProvider() *ui.ThemeProvider {
	return m.themeProvider
}

// renderHelpOverlay renders a help overlay on top of the current view
func (m Model) renderHelpOverlay(background string) string {
	// Build help content
	var help strings.Builder

	help.WriteString(m.styles.ViewTitle.Render("Keyboard Shortcuts"))
	help.WriteString("\n\n")

	// Global keys
	help.WriteString(m.styles.StatLabel.Render("Global:"))
	help.WriteString("\n")
	help.WriteString("  Tab/1-4    Switch views\n")
	help.WriteString("  ?          Toggle help\n")
	help.WriteString("  q          Quit\n")
	help.WriteString("\n")

	// View-specific keys
	switch m.activeTab {
	case TabEntries:
		help.WriteString(m.styles.StatLabel.Render("Entries:"))
		help.WriteString("\n")
		help.WriteString("  t          Today's entries\n")
		help.WriteString("  y          Yesterday's entries\n")
		help.WriteString("  w/W        This/Previous week\n")
		help.WriteString("  m/M        This/Previous month\n")
		help.WriteString("  j/k        Navigate up/down\n")
		help.WriteString("  n          New entry\n")
		help.WriteString("  e          Edit entry\n")
		help.WriteString("  d          Delete entry\n")
		help.WriteString("  s          Search entries\n")
		help.WriteString("  r          Refresh\n")
	case TabTimer:
		help.WriteString(m.styles.StatLabel.Render("Timer:"))
		help.WriteString("\n")
		help.WriteString("  s          Start timer\n")
		help.WriteString("  x          Stop timer\n")
		help.WriteString("  r          Refresh\n")
	case TabStats:
		help.WriteString(m.styles.StatLabel.Render("Stats:"))
		help.WriteString("\n")
		help.WriteString("  w          Weekly view\n")
		help.WriteString("  m          Monthly view\n")
		help.WriteString("  r          Refresh\n")
	case TabConfig:
		help.WriteString(m.styles.StatLabel.Render("Config:"))
		help.WriteString("\n")
		help.WriteString("  t/Enter    Open theme selector\n")
		help.WriteString("  j/k        Navigate themes\n")
		help.WriteString("  Enter      Select theme\n")
		help.WriteString("  Esc        Cancel\n")
	}

	help.WriteString("\n")
	help.WriteString(m.styles.StatLabel.Render("Press ? to close"))

	// Create a styled box for help
	helpBox := m.styles.Dialog.Render(help.String())

	// Center the help box (simple approach - just return it with the app style)
	return m.styles.App.Render(helpBox)
}

// Run starts the TUI application
func Run(services *service.Services) error {
	model := New(services)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
