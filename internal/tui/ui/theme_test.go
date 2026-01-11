package ui

import (
	"testing"
)

func TestNewThemeProvider_Default(t *testing.T) {
	tp := NewThemeProvider("")

	if tp == nil {
		t.Fatal("expected non-nil ThemeProvider")
	}

	// Should use default theme when empty string is passed
	if tp.CurrentName() != DefaultTheme {
		t.Errorf("expected default theme %q, got %q", DefaultTheme, tp.CurrentName())
	}
}

func TestNewThemeProvider_WithTheme(t *testing.T) {
	tp := NewThemeProvider("nord")

	if tp.CurrentName() != "nord" {
		t.Errorf("expected theme 'nord', got %q", tp.CurrentName())
	}
}

func TestNewThemeProvider_InvalidTheme(t *testing.T) {
	// Invalid theme should fall back to default
	tp := NewThemeProvider("nonexistent-theme-xyz")

	// Should still be usable
	if tp == nil {
		t.Fatal("expected non-nil ThemeProvider")
	}
}

func TestThemeProvider_SetTheme(t *testing.T) {
	tp := NewThemeProvider("")

	ok := tp.SetTheme("nord")
	if !ok {
		t.Error("expected SetTheme to return true for valid theme")
	}

	if tp.CurrentName() != "nord" {
		t.Errorf("expected theme 'nord', got %q", tp.CurrentName())
	}
}

func TestThemeProvider_SetTheme_Invalid(t *testing.T) {
	tp := NewThemeProvider("dracula")
	initialTheme := tp.CurrentName()

	ok := tp.SetTheme("nonexistent-theme-xyz")
	if ok {
		t.Error("expected SetTheme to return false for invalid theme")
	}

	// Theme should not change
	if tp.CurrentName() != initialTheme {
		t.Errorf("theme should not change after invalid SetTheme")
	}
}

func TestThemeProvider_NextTheme(t *testing.T) {
	tp := NewThemeProvider("dracula")
	initial := tp.CurrentName()

	next := tp.NextTheme()

	if next == initial {
		// With hundreds of themes, this is unlikely but possible if there's only one theme
		t.Log("NextTheme returned same theme (might be single theme)")
	}
	if tp.CurrentName() != next {
		t.Errorf("CurrentName() should match NextTheme() return value")
	}
}

func TestThemeProvider_PreviousTheme(t *testing.T) {
	tp := NewThemeProvider("dracula")
	initial := tp.CurrentName()

	prev := tp.PreviousTheme()

	if prev == initial {
		t.Log("PreviousTheme returned same theme")
	}
	if tp.CurrentName() != prev {
		t.Errorf("CurrentName() should match PreviousTheme() return value")
	}
}

func TestThemeProvider_AvailableThemes(t *testing.T) {
	tp := NewThemeProvider("")

	themes := tp.AvailableThemes()

	if len(themes) == 0 {
		t.Error("expected at least one available theme")
	}

	// Themes should be sorted
	for i := 1; i < len(themes); i++ {
		if themes[i] < themes[i-1] {
			t.Errorf("themes not sorted: %q < %q", themes[i], themes[i-1])
		}
	}

	// Should contain common themes
	found := false
	for _, theme := range themes {
		if theme == "dracula" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'dracula' in available themes")
	}
}

func TestThemeProvider_Styles(t *testing.T) {
	tp := NewThemeProvider("dracula")

	styles := tp.Styles()

	// Styles should be non-zero
	if styles.App.GetPaddingTop() == 0 && styles.App.GetPaddingBottom() == 0 {
		t.Error("expected App style to have padding")
	}
}

func TestThemeProvider_Registry(t *testing.T) {
	tp := NewThemeProvider("dracula")

	registry := tp.Registry()
	if registry == nil {
		t.Error("expected non-nil registry")
	}

	// Registry should provide colors
	color := registry.Purple()
	if color == nil {
		t.Error("expected non-nil Purple color from registry")
	}
}

func TestThemeProvider_CurrentDisplayName(t *testing.T) {
	tp := NewThemeProvider("dracula")

	displayName := tp.CurrentDisplayName()

	if displayName == "" {
		t.Error("expected non-empty display name")
	}
}
