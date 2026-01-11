package ui

import (
	"github.com/charmbracelet/lipgloss"
	tint "github.com/lrstanley/bubbletint"
)

// Styles contains all the styles used in the TUI
type Styles struct {
	// Base styles
	App lipgloss.Style

	// Tab bar
	TabBar       lipgloss.Style
	TabActive    lipgloss.Style
	TabInactive  lipgloss.Style
	TabSeparator lipgloss.Style

	// Content area
	Content   lipgloss.Style
	ViewTitle lipgloss.Style

	// Status bar
	StatusBar   lipgloss.Style
	StatusKey   lipgloss.Style
	StatusValue lipgloss.Style
	StatusHelp  lipgloss.Style

	// Entry list
	EntrySelected lipgloss.Style
	EntryNormal   lipgloss.Style
	EntryIndex    lipgloss.Style
	EntryTime     lipgloss.Style
	EntryDesc     lipgloss.Style
	EntryDuration lipgloss.Style
	EntryProject  lipgloss.Style
	EntryTag      lipgloss.Style

	// Timer
	TimerRunning lipgloss.Style
	TimerStopped lipgloss.Style
	TimerElapsed lipgloss.Style

	// Stats
	StatLabel lipgloss.Style
	StatValue lipgloss.Style

	// Help
	HelpKey  lipgloss.Style
	HelpDesc lipgloss.Style

	// Input
	Input        lipgloss.Style
	InputFocused lipgloss.Style

	// Dialog
	Dialog      lipgloss.Style
	DialogTitle lipgloss.Style

	// Errors and warnings
	Error   lipgloss.Style
	Warning lipgloss.Style
	Success lipgloss.Style
}

// DefaultStyles returns the default TUI styles
func DefaultStyles() Styles {
	// Color palette
	primary := lipgloss.Color("99")     // Purple
	secondary := lipgloss.Color("39")   // Cyan
	accent := lipgloss.Color("212")     // Pink
	muted := lipgloss.Color("240")      // Gray
	success := lipgloss.Color("82")     // Green
	warning := lipgloss.Color("214")    // Orange
	errorColor := lipgloss.Color("196") // Red

	return Styles{
		// Base styles
		App: lipgloss.NewStyle().Padding(1, 2),

		// Tab bar
		TabBar: lipgloss.NewStyle().
			MarginBottom(1).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(muted),
		TabActive: lipgloss.NewStyle().
			Foreground(primary).
			Bold(true).
			Padding(0, 2),
		TabInactive: lipgloss.NewStyle().
			Foreground(muted).
			Padding(0, 2),
		TabSeparator: lipgloss.NewStyle().
			Foreground(muted).
			SetString("|"),

		// Content area
		Content: lipgloss.NewStyle().
			Padding(0, 1),
		ViewTitle: lipgloss.NewStyle().
			Foreground(primary).
			Bold(true).
			MarginBottom(1),

		// Status bar
		StatusBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("236")).
			Padding(0, 1),
		StatusKey: lipgloss.NewStyle().
			Foreground(secondary).
			Bold(true),
		StatusValue: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		StatusHelp: lipgloss.NewStyle().
			Foreground(muted),

		// Entry list
		EntrySelected: lipgloss.NewStyle().
			Background(lipgloss.Color("237")).
			Bold(true),
		EntryNormal: lipgloss.NewStyle(),
		EntryIndex: lipgloss.NewStyle().
			Foreground(muted).
			Width(6),
		EntryTime: lipgloss.NewStyle().
			Foreground(secondary).
			Width(12),
		EntryDesc: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		EntryDuration: lipgloss.NewStyle().
			Foreground(accent).
			Width(10).
			Align(lipgloss.Right),
		EntryProject: lipgloss.NewStyle().
			Foreground(primary),
		EntryTag: lipgloss.NewStyle().
			Foreground(secondary),

		// Timer
		TimerRunning: lipgloss.NewStyle().
			Foreground(success).
			Bold(true),
		TimerStopped: lipgloss.NewStyle().
			Foreground(muted),
		TimerElapsed: lipgloss.NewStyle().
			Foreground(accent).
			Bold(true),

		// Stats
		StatLabel: lipgloss.NewStyle().
			Foreground(muted).
			Width(20),
		StatValue: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Bold(true),

		// Help
		HelpKey: lipgloss.NewStyle().
			Foreground(secondary).
			Bold(true),
		HelpDesc: lipgloss.NewStyle().
			Foreground(muted),

		// Input
		Input: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(muted).
			Padding(0, 1),
		InputFocused: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(primary).
			Padding(0, 1),

		// Dialog
		Dialog: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primary).
			Padding(1, 2).
			Width(50),
		DialogTitle: lipgloss.NewStyle().
			Foreground(primary).
			Bold(true).
			MarginBottom(1),

		// Errors and warnings
		Error: lipgloss.NewStyle().
			Foreground(errorColor),
		Warning: lipgloss.NewStyle().
			Foreground(warning),
		Success: lipgloss.NewStyle().
			Foreground(success),
	}
}

// NewStylesFromRegistry creates a Styles struct using colors from a bubbletint registry.
// This maps theme colors to semantic UI elements:
// - Primary: Purple (tabs, titles, projects)
// - Secondary: Cyan (times, tags, keys)
// - Accent: BrightPurple (durations, elapsed time)
// - Muted: BrightBlack (inactive elements, labels)
// - Success/Warning/Error: Green/Yellow/Red
func NewStylesFromRegistry(r *tint.Registry) Styles {
	// Get colors from registry (uses current theme)
	primary := r.Purple()
	secondary := r.Cyan()
	accent := r.BrightPurple()
	muted := r.BrightBlack()
	success := r.Green()
	warning := r.Yellow()
	errorColor := r.Red()
	fg := r.Fg()
	bg := r.Bg()

	return Styles{
		// Base styles
		App: lipgloss.NewStyle().Padding(1, 2),

		// Tab bar
		TabBar: lipgloss.NewStyle().
			MarginBottom(1).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(muted),
		TabActive: lipgloss.NewStyle().
			Foreground(primary).
			Bold(true).
			Padding(0, 2),
		TabInactive: lipgloss.NewStyle().
			Foreground(muted).
			Padding(0, 2),
		TabSeparator: lipgloss.NewStyle().
			Foreground(muted).
			SetString("|"),

		// Content area
		Content: lipgloss.NewStyle().
			Padding(0, 1),
		ViewTitle: lipgloss.NewStyle().
			Foreground(primary).
			Bold(true).
			MarginBottom(1),

		// Status bar
		StatusBar: lipgloss.NewStyle().
			Foreground(fg).
			Background(bg).
			Padding(0, 1),
		StatusKey: lipgloss.NewStyle().
			Foreground(secondary).
			Bold(true),
		StatusValue: lipgloss.NewStyle().
			Foreground(fg),
		StatusHelp: lipgloss.NewStyle().
			Foreground(muted),

		// Entry list
		EntrySelected: lipgloss.NewStyle().
			Background(muted).
			Bold(true),
		EntryNormal: lipgloss.NewStyle(),
		EntryIndex: lipgloss.NewStyle().
			Foreground(muted).
			Width(6),
		EntryTime: lipgloss.NewStyle().
			Foreground(secondary).
			Width(12),
		EntryDesc: lipgloss.NewStyle().
			Foreground(fg),
		EntryDuration: lipgloss.NewStyle().
			Foreground(accent).
			Width(10).
			Align(lipgloss.Right),
		EntryProject: lipgloss.NewStyle().
			Foreground(primary),
		EntryTag: lipgloss.NewStyle().
			Foreground(secondary),

		// Timer
		TimerRunning: lipgloss.NewStyle().
			Foreground(success).
			Bold(true),
		TimerStopped: lipgloss.NewStyle().
			Foreground(muted),
		TimerElapsed: lipgloss.NewStyle().
			Foreground(accent).
			Bold(true),

		// Stats
		StatLabel: lipgloss.NewStyle().
			Foreground(muted).
			Width(20),
		StatValue: lipgloss.NewStyle().
			Foreground(fg).
			Bold(true),

		// Help
		HelpKey: lipgloss.NewStyle().
			Foreground(secondary).
			Bold(true),
		HelpDesc: lipgloss.NewStyle().
			Foreground(muted),

		// Input
		Input: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(muted).
			Padding(0, 1),
		InputFocused: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(primary).
			Padding(0, 1),

		// Dialog
		Dialog: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primary).
			Padding(1, 2).
			Width(50),
		DialogTitle: lipgloss.NewStyle().
			Foreground(primary).
			Bold(true).
			MarginBottom(1),

		// Errors and warnings
		Error: lipgloss.NewStyle().
			Foreground(errorColor),
		Warning: lipgloss.NewStyle().
			Foreground(warning),
		Success: lipgloss.NewStyle().
			Foreground(success),
	}
}
