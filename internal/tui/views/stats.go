package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/xolan/did/internal/service"
	"github.com/xolan/did/internal/tui/ui"
)

// StatsModel is the model for the stats view
type StatsModel struct {
	services *service.Services
	styles   ui.Styles
	keys     ui.KeyMap

	// UI state
	width   int
	height  int
	result  *service.StatsResult
	loading bool
	err     error
	weekly  bool // true = weekly, false = monthly
}

// NewStatsModel creates a new stats view model
func NewStatsModel(services *service.Services, styles ui.Styles, keys ui.KeyMap) StatsModel {
	return StatsModel{
		services: services,
		styles:   styles,
		keys:     keys,
		weekly:   true,
	}
}

// statsLoadedMsg is sent when stats are loaded
type statsLoadedMsg struct {
	result *service.StatsResult
	err    error
}

// Init implements tea.Model
func (m StatsModel) Init() tea.Cmd {
	return m.loadStats()
}

// Update implements tea.Model
func (m StatsModel) Update(msg tea.Msg) (StatsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.ThisWeek):
			m.weekly = true
			return m, m.loadStats()
		case key.Matches(msg, m.keys.ThisMonth):
			m.weekly = false
			return m, m.loadStats()
		case key.Matches(msg, m.keys.Refresh):
			return m, m.loadStats()
		}

	case statsLoadedMsg:
		m.loading = false
		m.err = msg.err
		m.result = msg.result

	case ui.ThemeChangedMsg:
		m.styles = msg.Styles
		return m, nil
	}

	return m, nil
}

// View implements tea.Model
func (m StatsModel) View() string {
	var b strings.Builder

	title := "Weekly Statistics"
	if !m.weekly {
		title = "Monthly Statistics"
	}
	b.WriteString(m.styles.ViewTitle.Render(title))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString("Loading...")
		return b.String()
	}

	if m.err != nil {
		b.WriteString(m.styles.Error.Render(fmt.Sprintf("Error: %v", m.err)))
		return b.String()
	}

	if m.result == nil {
		b.WriteString("No data")
		return b.String()
	}

	// Statistics
	stats := m.result.Statistics
	b.WriteString(m.renderStatLine("Total time:", formatDuration(stats.TotalMinutes)))
	b.WriteString(m.renderStatLine("Total entries:", fmt.Sprintf("%d %s", stats.EntryCount, pluralize("entry", stats.EntryCount))))
	b.WriteString(m.renderStatLine("Days with work:", fmt.Sprintf("%d %s", stats.DaysWithEntries, pluralize("day", stats.DaysWithEntries))))
	b.WriteString(m.renderStatLine("Average per day:", formatDuration(int(stats.AverageMinutesPerDay))))

	// Comparison
	if m.result.Comparison != "" {
		b.WriteString("\n")
		b.WriteString(m.styles.StatLabel.Render("Comparison: "))
		b.WriteString(m.styles.StatValue.Render(m.result.Comparison))
		b.WriteString("\n")
	}

	// Project breakdown
	if len(m.result.ProjectStats) > 0 {
		b.WriteString("\n")
		b.WriteString(m.styles.ViewTitle.Render("By Project"))
		b.WriteString("\n")
		for _, ps := range m.result.ProjectStats {
			projectName := ps.Project
			if projectName != "(no project)" {
				projectName = "@" + projectName
			}
			line := fmt.Sprintf("  %-20s %10s  (%d %s)",
				projectName,
				formatDuration(ps.TotalMinutes),
				ps.EntryCount,
				pluralize("entry", ps.EntryCount))
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	return b.String()
}

// SetSize sets the view dimensions
func (m *StatsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// loadStats creates a command to load stats
func (m StatsModel) loadStats() tea.Cmd {
	return func() tea.Msg {
		var result *service.StatsResult
		var err error

		if m.weekly {
			result, err = m.services.Stats.Weekly()
		} else {
			result, err = m.services.Stats.Monthly()
		}

		return statsLoadedMsg{result: result, err: err}
	}
}

func (m StatsModel) renderStatLine(label, value string) string {
	return m.styles.StatLabel.Render(label) + " " + m.styles.StatValue.Render(value) + "\n"
}
