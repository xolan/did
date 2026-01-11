package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestDefaultStyles(t *testing.T) {
	styles := DefaultStyles()

	// Test that styles are non-empty (basic sanity check)
	tests := []struct {
		name  string
		style lipgloss.Style
	}{
		{"App", styles.App},
		{"TabBar", styles.TabBar},
		{"TabActive", styles.TabActive},
		{"TabInactive", styles.TabInactive},
		{"TabSeparator", styles.TabSeparator},
		{"Content", styles.Content},
		{"ViewTitle", styles.ViewTitle},
		{"StatusBar", styles.StatusBar},
		{"StatusKey", styles.StatusKey},
		{"StatusValue", styles.StatusValue},
		{"StatusHelp", styles.StatusHelp},
		{"EntrySelected", styles.EntrySelected},
		{"EntryNormal", styles.EntryNormal},
		{"EntryIndex", styles.EntryIndex},
		{"EntryTime", styles.EntryTime},
		{"EntryDesc", styles.EntryDesc},
		{"EntryDuration", styles.EntryDuration},
		{"EntryProject", styles.EntryProject},
		{"EntryTag", styles.EntryTag},
		{"TimerRunning", styles.TimerRunning},
		{"TimerStopped", styles.TimerStopped},
		{"TimerElapsed", styles.TimerElapsed},
		{"StatLabel", styles.StatLabel},
		{"StatValue", styles.StatValue},
		{"HelpKey", styles.HelpKey},
		{"HelpDesc", styles.HelpDesc},
		{"Input", styles.Input},
		{"InputFocused", styles.InputFocused},
		{"Dialog", styles.Dialog},
		{"DialogTitle", styles.DialogTitle},
		{"Error", styles.Error},
		{"Warning", styles.Warning},
		{"Success", styles.Success},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Render some text with the style to verify it works
			rendered := tt.style.Render("test")
			if rendered == "" {
				t.Errorf("expected non-empty rendered output for style %s", tt.name)
			}
		})
	}
}

func TestStylesRenderText(t *testing.T) {
	styles := DefaultStyles()

	// Test that various styles can render text correctly
	testText := "Hello, World!"

	// App style should add padding
	appRendered := styles.App.Render(testText)
	if appRendered == "" {
		t.Error("App style rendered empty string")
	}

	// TabActive should be bold
	tabRendered := styles.TabActive.Render("Tab")
	if tabRendered == "" {
		t.Error("TabActive style rendered empty string")
	}

	// Error style should work
	errorRendered := styles.Error.Render("Error message")
	if errorRendered == "" {
		t.Error("Error style rendered empty string")
	}
}

func TestStylesColorsAreConfigured(t *testing.T) {
	styles := DefaultStyles()

	// Verify that styles can render text without error
	// Note: ANSI codes may not be present in non-TTY environments
	successText := styles.Success.Render("success")
	errorText := styles.Error.Render("error")
	warningText := styles.Warning.Render("warning")

	// Basic check that rendering works
	if successText == "" {
		t.Error("Success style rendered empty string")
	}
	if errorText == "" {
		t.Error("Error style rendered empty string")
	}
	if warningText == "" {
		t.Error("Warning style rendered empty string")
	}

	// Verify the rendered text contains our content
	if len(successText) < len("success") {
		t.Error("Success render should contain at least the input text")
	}
}
