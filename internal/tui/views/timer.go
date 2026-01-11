package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/xolan/did/internal/service"
	"github.com/xolan/did/internal/timer"
	"github.com/xolan/did/internal/tui/ui"
)

// TimerModel is the model for the timer view
type TimerModel struct {
	services *service.Services
	styles   ui.Styles
	keys     ui.KeyMap

	// UI state
	width   int
	height  int
	status  *service.TimerStatus
	loading bool
	err     error

	// Input state for starting timer
	inputMode bool
	input     textinput.Model
}

// NewTimerModel creates a new timer view model
func NewTimerModel(services *service.Services, styles ui.Styles, keys ui.KeyMap) TimerModel {
	ti := textinput.New()
	ti.Placeholder = "Task description (@project #tag)..."
	ti.CharLimit = 200
	ti.Width = 50

	return TimerModel{
		services: services,
		styles:   styles,
		keys:     keys,
		input:    ti,
	}
}

// timerStatusMsg is sent when timer status is loaded
type timerStatusMsg struct {
	status *service.TimerStatus
	err    error
}

// timerTickMsg is sent every second to update elapsed time
type timerTickMsg time.Time

// Init implements tea.Model
func (m TimerModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadStatus(),
		m.tickTimer(),
	)
}

// Update implements tea.Model
func (m TimerModel) Update(msg tea.Msg) (TimerModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.inputMode {
			return m.handleInputMode(msg)
		}

		switch {
		case key.Matches(msg, m.keys.Start):
			// Only allow starting if no timer is running
			if m.status == nil || !m.status.Running {
				m.inputMode = true
				m.input.Focus()
				m.input.SetValue("")
				return m, textinput.Blink
			}
			return m, nil
		case key.Matches(msg, m.keys.Stop):
			if m.status != nil && m.status.Running {
				return m, m.stopTimer()
			}
			return m, nil
		case key.Matches(msg, m.keys.Refresh):
			return m, m.loadStatus()
		}

	case timerStatusMsg:
		m.loading = false
		m.err = msg.err
		m.status = msg.status
		m.inputMode = false
		return m, nil

	case timerTickMsg:
		if m.status != nil && m.status.Running {
			m.status.ElapsedTime = time.Since(m.status.State.StartedAt)
		}
		return m, m.tickTimer()

	case ui.ThemeChangedMsg:
		m.styles = msg.Styles
		return m, nil
	}

	// Update text input if in input mode
	if m.inputMode {
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleInputMode handles key events when in input mode
func (m TimerModel) handleInputMode(msg tea.KeyMsg) (TimerModel, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Select): // Enter
		desc := strings.TrimSpace(m.input.Value())
		if desc != "" {
			m.inputMode = false
			m.input.Blur()
			return m, m.startTimer(desc)
		}
		return m, nil
	case key.Matches(msg, m.keys.Back): // Escape
		m.inputMode = false
		m.input.Blur()
		m.input.SetValue("")
		return m, nil
	}

	// Pass other keys to text input
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View implements tea.Model
func (m TimerModel) View() string {
	var b strings.Builder

	b.WriteString(m.styles.ViewTitle.Render("Timer"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString("Loading...")
		return b.String()
	}

	if m.err != nil {
		b.WriteString(m.styles.Error.Render(fmt.Sprintf("Error: %v", m.err)))
		return b.String()
	}

	// Show input mode for starting timer
	if m.inputMode {
		b.WriteString(m.styles.StatLabel.Render("Start Timer"))
		b.WriteString("\n\n")
		b.WriteString(m.styles.StatLabel.Render("Description:"))
		b.WriteString("\n")
		b.WriteString(m.input.View())
		b.WriteString("\n\n")
		b.WriteString(m.styles.StatLabel.Render("Enter to start, Esc to cancel"))
		return b.String()
	}

	if m.status == nil || !m.status.Running {
		b.WriteString(m.styles.TimerStopped.Render("No timer running"))
		b.WriteString("\n\n")
		b.WriteString(m.styles.StatLabel.Render("Press 's' to start a new timer"))
		return b.String()
	}

	// Timer is running
	state := m.status.State
	b.WriteString(m.styles.TimerRunning.Render("● Timer Running"))
	b.WriteString("\n\n")

	// Description
	b.WriteString(m.styles.StatLabel.Render("Task:"))
	b.WriteString(" ")
	b.WriteString(m.styles.StatValue.Render(formatTimerDescription(state)))
	b.WriteString("\n")

	// Started at
	b.WriteString(m.styles.StatLabel.Render("Started:"))
	b.WriteString(" ")
	b.WriteString(m.styles.StatValue.Render(formatTimerStartTime(state.StartedAt)))
	b.WriteString("\n")

	// Elapsed time
	b.WriteString(m.styles.StatLabel.Render("Elapsed:"))
	b.WriteString(" ")
	b.WriteString(m.styles.TimerElapsed.Render(formatElapsedTime(m.status.ElapsedTime)))
	b.WriteString("\n\n")

	b.WriteString(m.styles.StatLabel.Render("Press 'x' to stop the timer"))

	return b.String()
}

// SetSize sets the view dimensions
func (m *TimerModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// loadStatus creates a command to load timer status
func (m TimerModel) loadStatus() tea.Cmd {
	return func() tea.Msg {
		status, err := m.services.Timer.Status()
		return timerStatusMsg{status: status, err: err}
	}
}

// startTimer creates a command to start the timer with the given description
func (m TimerModel) startTimer(description string) tea.Cmd {
	return func() tea.Msg {
		_, _, err := m.services.Timer.Start(description, false)
		if err != nil {
			return timerStatusMsg{err: err}
		}
		// Reload status after starting
		status, err := m.services.Timer.Status()
		return timerStatusMsg{status: status, err: err}
	}
}

// stopTimer creates a command to stop the timer
func (m TimerModel) stopTimer() tea.Cmd {
	return func() tea.Msg {
		_, _, err := m.services.Timer.Stop()
		if err != nil {
			return timerStatusMsg{err: err}
		}
		// Reload status after stopping
		status, err := m.services.Timer.Status()
		return timerStatusMsg{status: status, err: err}
	}
}

// tickTimer returns a command that sends a tick every second
func (m TimerModel) tickTimer() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return timerTickMsg(t)
	})
}

func formatTimerDescription(state *timer.TimerState) string {
	desc := state.Description
	if state.Project != "" {
		desc += " @" + state.Project
	}
	for _, tag := range state.Tags {
		desc += " #" + tag
	}
	return desc
}

func formatTimerStartTime(t time.Time) string {
	now := time.Now()
	if t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day() {
		return "today at " + t.Format("3:04 PM")
	}
	return t.Format("Mon Jan 2 at 3:04 PM")
}

func formatElapsedTime(d time.Duration) string {
	totalMinutes := int(d.Minutes())
	if totalMinutes < 60 {
		return fmt.Sprintf("%dm", totalMinutes)
	}
	hours := totalMinutes / 60
	mins := totalMinutes % 60
	if mins == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh %dm", hours, mins)
}

// IsInputMode returns true when the view is capturing keyboard input
func (m TimerModel) IsInputMode() bool {
	return m.inputMode
}
