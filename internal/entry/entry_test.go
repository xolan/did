package entry

import (
	"encoding/json"
	"testing"
)

func TestEntryBackwardCompatibility(t *testing.T) {
	// Test unmarshaling an old entry without project/tags fields
	oldJSON := `{"timestamp":"2024-01-01T10:00:00Z","description":"test entry","duration_minutes":60,"raw_input":"test entry for 1h"}`

	var e Entry
	if err := json.Unmarshal([]byte(oldJSON), &e); err != nil {
		t.Fatalf("failed to unmarshal old entry: %v", err)
	}

	if e.Description != "test entry" {
		t.Errorf("expected description 'test entry', got %q", e.Description)
	}
	if e.Project != "" {
		t.Errorf("expected empty project, got %q", e.Project)
	}
	if e.Tags != nil {
		t.Errorf("expected nil tags, got %v", e.Tags)
	}
}

func TestEntryWithProjectAndTags(t *testing.T) {
	// Test unmarshaling a new entry with project and tags
	newJSON := `{"timestamp":"2024-01-01T10:00:00Z","description":"test entry","duration_minutes":60,"raw_input":"test entry @acme #bugfix for 1h","project":"acme","tags":["bugfix","urgent"]}`

	var e Entry
	if err := json.Unmarshal([]byte(newJSON), &e); err != nil {
		t.Fatalf("failed to unmarshal new entry: %v", err)
	}

	if e.Project != "acme" {
		t.Errorf("expected project 'acme', got %q", e.Project)
	}
	if len(e.Tags) != 2 || e.Tags[0] != "bugfix" || e.Tags[1] != "urgent" {
		t.Errorf("expected tags [bugfix, urgent], got %v", e.Tags)
	}
}

func TestEntryOmitempty(t *testing.T) {
	// Test that empty project/tags are omitted from JSON output
	e := Entry{
		Description:     "test",
		DurationMinutes: 60,
	}

	data, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("failed to marshal entry: %v", err)
	}

	jsonStr := string(data)
	if contains(jsonStr, "project") {
		t.Errorf("expected project to be omitted, got %s", jsonStr)
	}
	if contains(jsonStr, "tags") {
		t.Errorf("expected tags to be omitted, got %s", jsonStr)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
