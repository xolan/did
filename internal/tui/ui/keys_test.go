package ui

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
)

func TestDefaultKeyMap(t *testing.T) {
	keys := DefaultKeyMap()

	// Test that all key bindings are properly configured
	tests := []struct {
		name    string
		binding key.Binding
	}{
		// Navigation
		{"Up", keys.Up},
		{"Down", keys.Down},
		{"Left", keys.Left},
		{"Right", keys.Right},

		// Tab navigation
		{"NextTab", keys.NextTab},
		{"PrevTab", keys.PrevTab},
		{"Tab1", keys.Tab1},
		{"Tab2", keys.Tab2},
		{"Tab3", keys.Tab3},
		{"Tab4", keys.Tab4},
		{"Tab5", keys.Tab5},

		// Actions
		{"Select", keys.Select},
		{"Back", keys.Back},
		{"Quit", keys.Quit},
		{"Help", keys.Help},
		{"Refresh", keys.Refresh},

		// Entry-specific
		{"New", keys.New},
		{"Edit", keys.Edit},
		{"Delete", keys.Delete},
		{"Undo", keys.Undo},
		{"Filter", keys.Filter},
		{"Search", keys.Search},

		// Timer-specific
		{"Start", keys.Start},
		{"Stop", keys.Stop},
		{"Cancel", keys.Cancel},

		// Date range shortcuts
		{"Today", keys.Today},
		{"Yesterday", keys.Yesterday},
		{"ThisWeek", keys.ThisWeek},
		{"PrevWeek", keys.PrevWeek},
		{"ThisMonth", keys.ThisMonth},
		{"PrevMonth", keys.PrevMonth},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check that the binding has keys defined
			if len(tt.binding.Keys()) == 0 {
				t.Errorf("expected keys for binding %s", tt.name)
			}
			// Check that help text is defined
			help := tt.binding.Help()
			if help.Key == "" {
				t.Errorf("expected help key for binding %s", tt.name)
			}
			if help.Desc == "" {
				t.Errorf("expected help description for binding %s", tt.name)
			}
		})
	}
}

func TestKeyBindingsMatch(t *testing.T) {
	keys := DefaultKeyMap()

	// Test that specific keys match their bindings
	tests := []struct {
		name    string
		binding key.Binding
		key     string
	}{
		{"Quit q", keys.Quit, "q"},
		{"Quit ctrl+c", keys.Quit, "ctrl+c"},
		{"Up k", keys.Up, "k"},
		{"Up arrow", keys.Up, "up"},
		{"Down j", keys.Down, "j"},
		{"Down arrow", keys.Down, "down"},
		{"Left h", keys.Left, "h"},
		{"Right l", keys.Right, "l"},
		{"Select enter", keys.Select, "enter"},
		{"Back esc", keys.Back, "esc"},
		{"Help ?", keys.Help, "?"},
		{"Tab1 1", keys.Tab1, "1"},
		{"Tab2 2", keys.Tab2, "2"},
		{"NextTab tab", keys.NextTab, "tab"},
		{"New n", keys.New, "n"},
		{"Edit e", keys.Edit, "e"},
		{"Delete d", keys.Delete, "d"},
		{"Start s", keys.Start, "s"},
		{"Stop x", keys.Stop, "x"},
		{"Search /", keys.Search, "/"},
		{"Today t", keys.Today, "t"},
		{"ThisWeek w", keys.ThisWeek, "w"},
		{"ThisMonth m", keys.ThisMonth, "m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found := false
			for _, k := range tt.binding.Keys() {
				if k == tt.key {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected binding %s to include key %s, got keys %v", tt.name, tt.key, tt.binding.Keys())
			}
		})
	}
}

func TestVimStyleNavigation(t *testing.T) {
	keys := DefaultKeyMap()

	// Verify vim-style hjkl navigation works
	vimKeys := []struct {
		binding key.Binding
		vimKey  string
	}{
		{keys.Up, "k"},
		{keys.Down, "j"},
		{keys.Left, "h"},
		{keys.Right, "l"},
	}

	for _, vk := range vimKeys {
		found := false
		for _, k := range vk.binding.Keys() {
			if k == vk.vimKey {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("vim key %s not found in binding keys %v", vk.vimKey, vk.binding.Keys())
		}
	}
}
