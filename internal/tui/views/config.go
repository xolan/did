package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/xolan/did/internal/config"
	"github.com/xolan/did/internal/service"
	"github.com/xolan/did/internal/tui/ui"
)

// ConfigModel is the model for the config view
type ConfigModel struct {
	services      *service.Services
	themeProvider *ui.ThemeProvider
	styles        ui.Styles
	keys          ui.KeyMap

	// UI state
	width     int
	height    int
	config    config.Config
	path      string
	exists    bool
	themeName string

	// Theme selector state
	selectingTheme bool
	themes         []string
	themeCursor    int
	themeOffset    int // For scrolling
}

// NewConfigModel creates a new config view model
func NewConfigModel(services *service.Services, themeProvider *ui.ThemeProvider, styles ui.Styles, keys ui.KeyMap) ConfigModel {
	themes := themeProvider.AvailableThemes()
	currentTheme := themeProvider.CurrentName()

	// Find cursor position for current theme
	cursor := 0
	for i, t := range themes {
		if t == currentTheme {
			cursor = i
			break
		}
	}

	return ConfigModel{
		services:      services,
		themeProvider: themeProvider,
		styles:        styles,
		keys:          keys,
		themes:        themes,
		themeCursor:   cursor,
	}
}

// Init implements tea.Model
func (m ConfigModel) Init() tea.Cmd {
	return m.loadConfig()
}

// configLoadedMsg is sent when config is loaded
type configLoadedMsg struct {
	config config.Config
	path   string
	exists bool
}

// maxVisibleThemes is the maximum number of themes to show at once
const maxVisibleThemes = 10

// Update implements tea.Model
func (m ConfigModel) Update(msg tea.Msg) (ConfigModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.selectingTheme {
			return m.handleThemeSelection(msg)
		}

		// Open theme selector with Enter or 't'
		if key.Matches(msg, m.keys.Select) || msg.String() == "t" {
			m.selectingTheme = true
			// Center the current theme in view
			m.updateThemeOffset()
			return m, nil
		}

	case configLoadedMsg:
		m.config = msg.config
		m.path = msg.path
		m.exists = msg.exists
		m.themeName = msg.config.Theme
		if m.themeName == "" {
			m.themeName = ui.DefaultTheme
		}
		// Update cursor to match loaded theme
		for i, t := range m.themes {
			if t == m.themeName {
				m.themeCursor = i
				break
			}
		}

	case ui.ThemeChangedMsg:
		m.styles = msg.Styles
		m.themeName = msg.ThemeName
		return m, nil
	}

	return m, nil
}

// handleThemeSelection handles keys when theme selector is open
func (m ConfigModel) handleThemeSelection(msg tea.KeyMsg) (ConfigModel, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.themeCursor > 0 {
			m.themeCursor--
			m.updateThemeOffset()
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.themeCursor < len(m.themes)-1 {
			m.themeCursor++
			m.updateThemeOffset()
		}
		return m, nil

	case key.Matches(msg, m.keys.Select):
		// Select theme and close selector
		selectedTheme := m.themes[m.themeCursor]
		m.selectingTheme = false
		return m, m.requestThemeChange(selectedTheme)

	case key.Matches(msg, m.keys.Back):
		// Close selector without changing
		m.selectingTheme = false
		// Reset cursor to current theme
		for i, t := range m.themes {
			if t == m.themeName {
				m.themeCursor = i
				break
			}
		}
		return m, nil
	}

	return m, nil
}

// updateThemeOffset adjusts scroll offset to keep cursor visible
func (m *ConfigModel) updateThemeOffset() {
	// Ensure cursor is within visible range
	if m.themeCursor < m.themeOffset {
		m.themeOffset = m.themeCursor
	} else if m.themeCursor >= m.themeOffset+maxVisibleThemes {
		m.themeOffset = m.themeCursor - maxVisibleThemes + 1
	}
}

// requestThemeChange creates a command to request a theme change by name
func (m ConfigModel) requestThemeChange(themeName string) tea.Cmd {
	return func() tea.Msg {
		return ui.ThemeChangeRequestMsg{ThemeName: themeName}
	}
}

// View implements tea.Model
func (m ConfigModel) View() string {
	var b strings.Builder

	b.WriteString(m.styles.ViewTitle.Render("Configuration"))
	b.WriteString("\n\n")

	// Config file path
	b.WriteString(m.styles.StatLabel.Render("Config file:"))
	b.WriteString(" ")
	b.WriteString(m.styles.StatValue.Render(m.path))
	b.WriteString("\n")

	// Status
	b.WriteString(m.styles.StatLabel.Render("Status:"))
	b.WriteString(" ")
	if m.exists {
		b.WriteString(m.styles.Success.Render("File exists"))
	} else {
		b.WriteString(m.styles.Warning.Render("Using defaults (no config file)"))
	}
	b.WriteString("\n\n")

	// Config values
	b.WriteString(strings.Repeat("─", min(50, m.width)))
	b.WriteString("\n\n")

	b.WriteString(m.renderConfigLine("week_start_day", m.config.WeekStartDay))
	b.WriteString(m.renderConfigLine("timezone", m.config.Timezone))

	// Theme with selector
	if m.selectingTheme {
		b.WriteString(m.renderThemeSelector())
	} else {
		b.WriteString(m.renderConfigLine("theme", m.themeName))
		b.WriteString("\n")
		b.WriteString(m.styles.StatLabel.Render("Press Enter or 't' to change theme"))
	}

	return b.String()
}

// renderThemeSelector renders the theme selection list
func (m ConfigModel) renderThemeSelector() string {
	var b strings.Builder

	b.WriteString(m.styles.StatLabel.Render("theme:"))
	b.WriteString(" ")
	b.WriteString(m.styles.StatValue.Render("Select a theme"))
	b.WriteString("\n\n")

	// Calculate visible range
	endIdx := m.themeOffset + maxVisibleThemes
	if endIdx > len(m.themes) {
		endIdx = len(m.themes)
	}

	// Show scroll indicator at top if needed
	if m.themeOffset > 0 {
		b.WriteString(m.styles.StatLabel.Render("  ↑ more themes above"))
		b.WriteString("\n")
	}

	// Render visible themes
	for i := m.themeOffset; i < endIdx; i++ {
		theme := m.themes[i]
		if i == m.themeCursor {
			// Highlighted/selected theme
			b.WriteString(m.styles.EntrySelected.Render("▸ " + theme))
			if theme == m.themeName {
				b.WriteString(m.styles.Success.Render(" (current)"))
			}
		} else {
			b.WriteString("  ")
			if theme == m.themeName {
				b.WriteString(m.styles.Success.Render(theme + " (current)"))
			} else {
				b.WriteString(m.styles.StatValue.Render(theme))
			}
		}
		b.WriteString("\n")
	}

	// Show scroll indicator at bottom if needed
	if endIdx < len(m.themes) {
		b.WriteString(m.styles.StatLabel.Render("  ↓ more themes below"))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.styles.StatLabel.Render("↑/↓ navigate  Enter select  Esc cancel"))

	return b.String()
}

// SetSize sets the view dimensions
func (m *ConfigModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// loadConfig creates a command to load config
func (m ConfigModel) loadConfig() tea.Cmd {
	return func() tea.Msg {
		cfg := m.services.Config.Get()
		path := m.services.Config.GetPath()
		exists := m.services.Config.Exists()
		return configLoadedMsg{
			config: cfg,
			path:   path,
			exists: exists,
		}
	}
}

func (m ConfigModel) renderConfigLine(key, value string) string {
	return m.styles.StatLabel.Render(key+":") + " " + m.styles.StatValue.Render(value) + "\n"
}
