package ui

// ThemeChangeRequestMsg is sent when a theme change is requested.
type ThemeChangeRequestMsg struct {
	ThemeName string
}

// ThemeChangedMsg is broadcast to all views when the theme changes.
type ThemeChangedMsg struct {
	ThemeName string
	Styles    Styles
}
