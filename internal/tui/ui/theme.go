package ui

import (
	"sort"

	tint "github.com/lrstanley/bubbletint"
)

// DefaultTheme is the default theme used when no theme is configured
const DefaultTheme = "dracula"

// ThemeProvider manages TUI themes using bubbletint
type ThemeProvider struct {
	registry *tint.Registry
}

// NewThemeProvider creates a new ThemeProvider with the specified initial theme.
// If initialTheme is empty, DefaultTheme is used.
// If the specified theme doesn't exist, the default theme is used.
func NewThemeProvider(initialTheme string) *ThemeProvider {
	// Get all available tints
	allTints := tint.DefaultTints()

	// Find the default tint
	var defaultTint tint.Tint
	for _, t := range allTints {
		if t.ID() == DefaultTheme {
			defaultTint = t
			break
		}
	}

	// Fallback to first tint if default not found
	if defaultTint == nil && len(allTints) > 0 {
		defaultTint = allTints[0]
	}

	// Create registry with all tints
	registry := tint.NewRegistry(defaultTint, allTints...)

	// Set initial theme if specified
	if initialTheme != "" {
		registry.SetTintID(initialTheme)
	}

	return &ThemeProvider{
		registry: registry,
	}
}

// SetTheme sets the current theme by name.
// Returns true if the theme was found and set, false otherwise.
func (tp *ThemeProvider) SetTheme(name string) bool {
	return tp.registry.SetTintID(name)
}

// NextTheme cycles to the next theme.
// Returns the name of the new current theme.
func (tp *ThemeProvider) NextTheme() string {
	tp.registry.NextTint()
	return tp.registry.ID()
}

// PreviousTheme cycles to the previous theme.
// Returns the name of the new current theme.
func (tp *ThemeProvider) PreviousTheme() string {
	tp.registry.PreviousTint()
	return tp.registry.ID()
}

// CurrentName returns the name of the current theme.
func (tp *ThemeProvider) CurrentName() string {
	return tp.registry.ID()
}

// CurrentDisplayName returns the display name of the current theme.
func (tp *ThemeProvider) CurrentDisplayName() string {
	return tp.registry.DisplayName()
}

// AvailableThemes returns a sorted list of all available theme names.
func (tp *ThemeProvider) AvailableThemes() []string {
	ids := tp.registry.TintIDs()
	sort.Strings(ids)
	return ids
}

// Registry returns the underlying bubbletint registry for direct color access.
func (tp *ThemeProvider) Registry() *tint.Registry {
	return tp.registry
}

// Styles returns a Styles struct configured for the current theme.
func (tp *ThemeProvider) Styles() Styles {
	return NewStylesFromRegistry(tp.registry)
}
