package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/xolan/did/internal/service"
	"github.com/xolan/did/internal/tui/ui"
)

// entryMode represents the current mode of the entries view
type entryMode int

const (
	entryModeNormal entryMode = iota
	entryModeAdd
	entryModeEdit
	entryModeDelete
	entryModeSearch
)

// EntriesModel is the model for the entries view
type EntriesModel struct {
	services *service.Services
	styles   ui.Styles
	keys     ui.KeyMap

	// UI state
	width   int
	height  int
	cursor  int
	entries []service.IndexedEntry
	period  string
	total   int
	loading bool
	err     error

	// Date range
	dateRange service.DateRangeSpec

	// Input mode state
	mode          entryMode
	descInput     textinput.Model
	durationInput textinput.Model
	focusedInput  int // 0 = description, 1 = duration
	editIndex     int // Index of entry being edited

	// Search mode state
	searchInput   textinput.Model
	searchResults []service.IndexedEntry
	searchCursor  int
	searched      bool
}

// NewEntriesModel creates a new entries view model
func NewEntriesModel(services *service.Services, styles ui.Styles, keys ui.KeyMap) EntriesModel {
	descInput := textinput.New()
	descInput.Placeholder = "Task description (@project #tag)..."
	descInput.CharLimit = 200
	descInput.Width = 50

	durationInput := textinput.New()
	durationInput.Placeholder = "Duration (e.g., 1h30m, 45m, 2h)..."
	durationInput.CharLimit = 20
	durationInput.Width = 20

	searchInput := textinput.New()
	searchInput.Placeholder = "Search entries..."
	searchInput.CharLimit = 100
	searchInput.Width = 40

	return EntriesModel{
		services:      services,
		styles:        styles,
		keys:          keys,
		dateRange:     service.DateRangeSpec{Type: service.DateRangeToday},
		descInput:     descInput,
		durationInput: durationInput,
		searchInput:   searchInput,
	}
}

// entriesLoadedMsg is sent when entries are loaded
type entriesLoadedMsg struct {
	entries []service.IndexedEntry
	period  string
	total   int
	err     error
}

// searchResultsMsg is sent when search results are loaded
type searchResultsMsg struct {
	results []service.IndexedEntry
	err     error
}

// Init implements tea.Model
func (m EntriesModel) Init() tea.Cmd {
	return m.loadEntries()
}

// Update implements tea.Model
func (m EntriesModel) Update(msg tea.Msg) (EntriesModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle input modes
		switch m.mode {
		case entryModeAdd, entryModeEdit:
			return m.handleInputMode(msg)
		case entryModeDelete:
			return m.handleDeleteMode(msg)
		case entryModeSearch:
			return m.handleSearchMode(msg)
		}

		// Normal mode key handling
		switch {
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			}
			return m, nil
		case key.Matches(msg, m.keys.Today):
			m.dateRange = service.DateRangeSpec{Type: service.DateRangeToday}
			return m, m.loadEntries()
		case key.Matches(msg, m.keys.Yesterday):
			m.dateRange = service.DateRangeSpec{Type: service.DateRangeYesterday}
			return m, m.loadEntries()
		case key.Matches(msg, m.keys.ThisWeek):
			m.dateRange = service.DateRangeSpec{Type: service.DateRangeThisWeek}
			return m, m.loadEntries()
		case key.Matches(msg, m.keys.PrevWeek):
			m.dateRange = service.DateRangeSpec{Type: service.DateRangePrevWeek}
			return m, m.loadEntries()
		case key.Matches(msg, m.keys.ThisMonth):
			m.dateRange = service.DateRangeSpec{Type: service.DateRangeThisMonth}
			return m, m.loadEntries()
		case key.Matches(msg, m.keys.Refresh):
			return m, m.loadEntries()
		case key.Matches(msg, m.keys.New):
			m.mode = entryModeAdd
			m.descInput.SetValue("")
			m.durationInput.SetValue("")
			m.focusedInput = 0
			m.descInput.Focus()
			return m, textinput.Blink
		case key.Matches(msg, m.keys.Edit):
			if len(m.entries) > 0 && m.cursor < len(m.entries) {
				m.mode = entryModeEdit
				entry := m.entries[m.cursor].Entry
				m.editIndex = m.entries[m.cursor].ActiveIndex
				m.descInput.SetValue(entry.Description)
				m.durationInput.SetValue(formatDuration(entry.DurationMinutes))
				m.focusedInput = 0
				m.descInput.Focus()
				return m, textinput.Blink
			}
			return m, nil
		case key.Matches(msg, m.keys.Delete):
			if len(m.entries) > 0 && m.cursor < len(m.entries) {
				m.mode = entryModeDelete
			}
			return m, nil
		case key.Matches(msg, m.keys.Search), key.Matches(msg, m.keys.Start):
			// Use both '/' (Search) and 's' (Start) for search in entries view
			m.mode = entryModeSearch
			m.searchInput.SetValue("")
			m.searchInput.Focus()
			m.searched = false
			m.searchResults = nil
			m.searchCursor = 0
			return m, textinput.Blink
		}

	case entriesLoadedMsg:
		m.loading = false
		m.err = msg.err
		m.mode = entryModeNormal
		if msg.err == nil {
			m.entries = msg.entries
			m.period = msg.period
			m.total = msg.total
			if m.cursor >= len(m.entries) {
				m.cursor = max(0, len(m.entries)-1)
			}
		}

	case ui.ThemeChangedMsg:
		m.styles = msg.Styles
		return m, nil

	case searchResultsMsg:
		m.searched = true
		m.err = msg.err
		if msg.err == nil {
			m.searchResults = msg.results
			m.searchCursor = 0
		}
		return m, nil
	}

	// Update text inputs if in input mode
	if m.mode == entryModeAdd || m.mode == entryModeEdit {
		if m.focusedInput == 0 {
			m.descInput, cmd = m.descInput.Update(msg)
		} else {
			m.durationInput, cmd = m.durationInput.Update(msg)
		}
		return m, cmd
	}

	// Update search input if in search mode
	if m.mode == entryModeSearch && m.searchInput.Focused() {
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleInputMode handles key events when in add/edit mode
func (m EntriesModel) handleInputMode(msg tea.KeyMsg) (EntriesModel, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Select): // Enter
		desc := strings.TrimSpace(m.descInput.Value())
		dur := strings.TrimSpace(m.durationInput.Value())
		if desc != "" && dur != "" {
			m.descInput.Blur()
			m.durationInput.Blur()
			if m.mode == entryModeAdd {
				return m, m.addEntry(desc, dur)
			}
			return m, m.editEntry(m.editIndex, desc, dur)
		}
		return m, nil
	case key.Matches(msg, m.keys.Back): // Escape
		m.mode = entryModeNormal
		m.descInput.Blur()
		m.durationInput.Blur()
		return m, nil
	case msg.String() == "tab":
		// Switch between inputs
		if m.focusedInput == 0 {
			m.focusedInput = 1
			m.descInput.Blur()
			m.durationInput.Focus()
		} else {
			m.focusedInput = 0
			m.durationInput.Blur()
			m.descInput.Focus()
		}
		return m, textinput.Blink
	}

	// Pass other keys to focused input
	var cmd tea.Cmd
	if m.focusedInput == 0 {
		m.descInput, cmd = m.descInput.Update(msg)
	} else {
		m.durationInput, cmd = m.durationInput.Update(msg)
	}
	return m, cmd
}

// handleDeleteMode handles key events when in delete confirmation mode
func (m EntriesModel) handleDeleteMode(msg tea.KeyMsg) (EntriesModel, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if m.cursor < len(m.entries) {
			index := m.entries[m.cursor].ActiveIndex
			m.mode = entryModeNormal
			return m, m.deleteEntry(index)
		}
	case "n", "N", "esc":
		m.mode = entryModeNormal
	}
	return m, nil
}

// handleSearchMode handles key events when in search mode
func (m EntriesModel) handleSearchMode(msg tea.KeyMsg) (EntriesModel, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Select): // Enter
		if m.searchInput.Focused() {
			query := strings.TrimSpace(m.searchInput.Value())
			if query != "" {
				m.searchInput.Blur()
				return m, m.searchEntries(query)
			}
		}
		return m, nil
	case key.Matches(msg, m.keys.Back): // Escape
		m.mode = entryModeNormal
		m.searchInput.Blur()
		m.searched = false
		m.searchResults = nil
		return m, nil
	case key.Matches(msg, m.keys.Up):
		if !m.searchInput.Focused() && m.searchCursor > 0 {
			m.searchCursor--
		}
		return m, nil
	case key.Matches(msg, m.keys.Down):
		if !m.searchInput.Focused() && m.searchCursor < len(m.searchResults)-1 {
			m.searchCursor++
		}
		return m, nil
	case msg.String() == "/":
		// Re-focus search input
		m.searchInput.Focus()
		return m, textinput.Blink
	}

	// Pass other keys to search input if focused
	if m.searchInput.Focused() {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View implements tea.Model
func (m EntriesModel) View() string {
	var b strings.Builder

	// Handle special modes
	switch m.mode {
	case entryModeAdd:
		return m.renderAddForm()
	case entryModeEdit:
		return m.renderEditForm()
	case entryModeDelete:
		return m.renderDeleteConfirm()
	case entryModeSearch:
		return m.renderSearchView()
	}

	// Title with period
	title := fmt.Sprintf("Entries for %s", m.period)
	b.WriteString(m.styles.ViewTitle.Render(title))
	b.WriteString("\n")

	if m.loading {
		b.WriteString("Loading...")
		return b.String()
	}

	if m.err != nil {
		b.WriteString(m.styles.Error.Render(fmt.Sprintf("Error: %v", m.err)))
		return b.String()
	}

	if len(m.entries) == 0 {
		b.WriteString(m.styles.StatLabel.Render("No entries found"))
		b.WriteString("\n\n")
		b.WriteString(m.styles.StatLabel.Render("Press 'n' to add a new entry"))
		return b.String()
	}

	// Render entries using shared renderer
	b.WriteString(RenderEntryList(m.entries, m.styles, EntryRenderOptions{
		ShowDate: m.isMultiDayRange(),
		Width:    m.width,
		Cursor:   m.cursor,
	}))

	// Total
	b.WriteString(strings.Repeat("─", min(50, m.width)))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Total: %s (%d %s)",
		formatDuration(m.total),
		len(m.entries),
		pluralize("entry", len(m.entries))))

	return b.String()
}

// renderAddForm renders the add entry form
func (m EntriesModel) renderAddForm() string {
	var b strings.Builder
	b.WriteString(m.styles.ViewTitle.Render("New Entry"))
	b.WriteString("\n\n")
	b.WriteString(m.renderEntryForm())
	return b.String()
}

// renderEditForm renders the edit entry form
func (m EntriesModel) renderEditForm() string {
	var b strings.Builder
	b.WriteString(m.styles.ViewTitle.Render("Edit Entry"))
	b.WriteString("\n\n")
	b.WriteString(m.renderEntryForm())
	return b.String()
}

// renderEntryForm renders the common entry form fields
func (m EntriesModel) renderEntryForm() string {
	var b strings.Builder

	// Description input
	descLabel := "Description:"
	if m.focusedInput == 0 {
		descLabel = "▸ Description:"
	}
	b.WriteString(m.styles.StatLabel.Render(descLabel))
	b.WriteString("\n")
	b.WriteString(m.descInput.View())
	b.WriteString("\n\n")

	// Duration input
	durLabel := "Duration:"
	if m.focusedInput == 1 {
		durLabel = "▸ Duration:"
	}
	b.WriteString(m.styles.StatLabel.Render(durLabel))
	b.WriteString("\n")
	b.WriteString(m.durationInput.View())
	b.WriteString("\n\n")

	b.WriteString(m.styles.StatLabel.Render("Tab to switch fields, Enter to save, Esc to cancel"))
	return b.String()
}

// renderDeleteConfirm renders the delete confirmation dialog
func (m EntriesModel) renderDeleteConfirm() string {
	var b strings.Builder
	b.WriteString(m.styles.ViewTitle.Render("Delete Entry"))
	b.WriteString("\n\n")

	if m.cursor < len(m.entries) {
		entry := m.entries[m.cursor].Entry
		b.WriteString(m.styles.Warning.Render("Are you sure you want to delete this entry?"))
		b.WriteString("\n\n")
		b.WriteString(m.styles.StatLabel.Render("Description: "))
		b.WriteString(m.styles.StatValue.Render(entry.Description))
		b.WriteString("\n")
		b.WriteString(m.styles.StatLabel.Render("Duration: "))
		b.WriteString(m.styles.StatValue.Render(formatDuration(entry.DurationMinutes)))
		b.WriteString("\n\n")
	}

	b.WriteString(m.styles.StatLabel.Render("Press Y to confirm, N or Esc to cancel"))
	return b.String()
}

// renderSearchView renders the search interface
func (m EntriesModel) renderSearchView() string {
	var b strings.Builder

	b.WriteString(m.styles.ViewTitle.Render("Search Entries"))
	b.WriteString("\n\n")

	// Search input
	b.WriteString(m.searchInput.View())
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(m.styles.Error.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n\n")
	}

	if !m.searched {
		b.WriteString(m.styles.StatLabel.Render("Enter a search term and press Enter"))
		b.WriteString("\n\n")
		b.WriteString(m.styles.StatLabel.Render("Press Esc to return to entries"))
		return b.String()
	}

	if len(m.searchResults) == 0 {
		b.WriteString(m.styles.StatLabel.Render("No results found"))
		b.WriteString("\n\n")
		b.WriteString(m.styles.StatLabel.Render("Press / to search again, Esc to return"))
		return b.String()
	}

	// Results count
	b.WriteString(fmt.Sprintf("Found %d %s:\n\n", len(m.searchResults), pluralize("result", len(m.searchResults))))

	// Render results using shared renderer (always show date)
	b.WriteString(RenderEntryList(m.searchResults, m.styles, EntryRenderOptions{
		ShowDate: true,
		Width:    m.width,
		Cursor:   m.searchCursor,
	}))

	b.WriteString("\n")
	b.WriteString(m.styles.StatLabel.Render("j/k navigate  / search again  Esc return"))

	return b.String()
}

// SetSize sets the view dimensions
func (m *EntriesModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// GetDateRangeType returns the current date range type for testing
func (m EntriesModel) GetDateRangeType() service.DateRange {
	return m.dateRange.Type
}

// isMultiDayRange returns true if the current date range spans multiple days
func (m EntriesModel) isMultiDayRange() bool {
	switch m.dateRange.Type {
	case service.DateRangeToday, service.DateRangeYesterday:
		return false
	default:
		// Week, month, last N days, custom ranges all span multiple days
		return true
	}
}

// loadEntries creates a command to load entries
func (m EntriesModel) loadEntries() tea.Cmd {
	return func() tea.Msg {
		result, err := m.services.Entry.List(m.dateRange, nil)
		if err != nil {
			return entriesLoadedMsg{err: err}
		}
		return entriesLoadedMsg{
			entries: result.Entries,
			period:  result.Period,
			total:   result.Total,
		}
	}
}

// addEntry creates a command to add a new entry
func (m EntriesModel) addEntry(description, duration string) tea.Cmd {
	return func() tea.Msg {
		// Format: "description for duration"
		input := description + " for " + duration
		_, err := m.services.Entry.Create(input)
		if err != nil {
			return entriesLoadedMsg{err: err}
		}
		// Reload entries after adding
		result, err := m.services.Entry.List(m.dateRange, nil)
		if err != nil {
			return entriesLoadedMsg{err: err}
		}
		return entriesLoadedMsg{
			entries: result.Entries,
			period:  result.Period,
			total:   result.Total,
		}
	}
}

// editEntry creates a command to edit an existing entry
func (m EntriesModel) editEntry(index int, description, duration string) tea.Cmd {
	return func() tea.Msg {
		_, err := m.services.Entry.Edit(index, description, duration)
		if err != nil {
			return entriesLoadedMsg{err: err}
		}
		// Reload entries after editing
		result, err := m.services.Entry.List(m.dateRange, nil)
		if err != nil {
			return entriesLoadedMsg{err: err}
		}
		return entriesLoadedMsg{
			entries: result.Entries,
			period:  result.Period,
			total:   result.Total,
		}
	}
}

// deleteEntry creates a command to delete an entry
func (m EntriesModel) deleteEntry(index int) tea.Cmd {
	return func() tea.Msg {
		_, err := m.services.Entry.Delete(index)
		if err != nil {
			return entriesLoadedMsg{err: err}
		}
		// Reload entries after deleting
		result, err := m.services.Entry.List(m.dateRange, nil)
		if err != nil {
			return entriesLoadedMsg{err: err}
		}
		return entriesLoadedMsg{
			entries: result.Entries,
			period:  result.Period,
			total:   result.Total,
		}
	}
}

// searchEntries creates a command to search entries
func (m EntriesModel) searchEntries(query string) tea.Cmd {
	return func() tea.Msg {
		result, err := m.services.Search.Search(query, nil, nil)
		if err != nil {
			return searchResultsMsg{err: err}
		}
		return searchResultsMsg{results: result.Entries}
	}
}

// IsInputMode returns true when the view is capturing keyboard input
func (m EntriesModel) IsInputMode() bool {
	return m.mode == entryModeAdd || m.mode == entryModeEdit
}
