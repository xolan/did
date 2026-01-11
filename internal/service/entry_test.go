package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xolan/did/internal/config"
	"github.com/xolan/did/internal/entry"
	"github.com/xolan/did/internal/filter"
)

func TestNewEntryService(t *testing.T) {
	svc := NewEntryService("/tmp/test.jsonl", config.DefaultConfig())
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.storagePath != "/tmp/test.jsonl" {
		t.Errorf("expected storagePath '/tmp/test.jsonl', got %q", svc.storagePath)
	}
}

func TestEntryService_Create(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewEntryService(storagePath, config.DefaultConfig())

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:    "valid entry",
			input:   "fix bug for 2h",
			wantErr: nil,
		},
		{
			name:    "entry with project",
			input:   "fix bug @acme for 30m",
			wantErr: nil,
		},
		{
			name:    "entry with tags",
			input:   "code review #review #urgent for 1h",
			wantErr: nil,
		},
		{
			name:    "missing duration",
			input:   "fix bug",
			wantErr: ErrMissingDuration,
		},
		{
			name:    "empty description",
			input:   " for 2h",
			wantErr: ErrEmptyDescription,
		},
		{
			name:    "only project/tags",
			input:   "@acme #tag for 2h",
			wantErr: ErrEmptyDescription,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := svc.Create(tt.input)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.wantErr)
				} else if err != tt.wantErr {
					// Check if the error message matches for wrapped errors
					if err.Error() != tt.wantErr.Error() && !containsError(err, tt.wantErr) {
						t.Errorf("expected error %v, got %v", tt.wantErr, err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if e == nil {
					t.Error("expected entry, got nil")
				}
			}
		})
	}
}

func TestEntryService_Create_InvalidDuration(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewEntryService(storagePath, config.DefaultConfig())

	_, err := svc.Create("fix bug for invalid")
	if err == nil {
		t.Error("expected error for invalid duration")
	}
}

func TestEntryService_CreateFromParts(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewEntryService(storagePath, config.DefaultConfig())

	tests := []struct {
		name        string
		description string
		duration    int
		project     string
		tags        []string
		wantErr     bool
	}{
		{
			name:        "valid entry",
			description: "fix bug",
			duration:    120,
			project:     "",
			tags:        nil,
			wantErr:     false,
		},
		{
			name:        "with project and tags",
			description: "fix bug",
			duration:    60,
			project:     "acme",
			tags:        []string{"urgent"},
			wantErr:     false,
		},
		{
			name:        "empty description",
			description: "",
			duration:    60,
			project:     "",
			tags:        nil,
			wantErr:     true,
		},
		{
			name:        "zero duration",
			description: "fix bug",
			duration:    0,
			project:     "",
			tags:        nil,
			wantErr:     true,
		},
		{
			name:        "negative duration",
			description: "fix bug",
			duration:    -10,
			project:     "",
			tags:        nil,
			wantErr:     true,
		},
		{
			name:        "excessive duration",
			description: "fix bug",
			duration:    entry.MaxDurationMinutes + 1,
			project:     "",
			tags:        nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := svc.CreateFromParts(tt.description, tt.duration, tt.project, tt.tags)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if e == nil {
					t.Error("expected entry, got nil")
				}
				if e != nil && e.Description != tt.description {
					t.Errorf("expected description %q, got %q", tt.description, e.Description)
				}
			}
		})
	}
}

func TestEntryService_List(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewEntryService(storagePath, config.DefaultConfig())

	// Create some entries
	_, _ = svc.Create("task 1 for 1h")
	_, _ = svc.Create("task 2 @project for 30m")
	_, _ = svc.Create("task 3 #tag for 45m")

	// List today's entries
	result, err := svc.List(DateRangeSpec{Type: DateRangeToday}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(result.Entries))
	}
	if result.Total != 135 { // 60 + 30 + 45
		t.Errorf("expected total 135 minutes, got %d", result.Total)
	}
}

func TestEntryService_List_WithFilter(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewEntryService(storagePath, config.DefaultConfig())

	// Create entries with different projects
	_, _ = svc.Create("task 1 @acme for 1h")
	_, _ = svc.Create("task 2 @other for 30m")
	_, _ = svc.Create("task 3 @acme for 45m")

	// Filter by project
	f := filter.NewFilter("", "acme", nil)
	result, err := svc.List(DateRangeSpec{Type: DateRangeToday}, f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(result.Entries))
	}
}

func TestEntryService_List_DateRanges(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewEntryService(storagePath, config.DefaultConfig())

	// Test all date range types
	ranges := []DateRangeSpec{
		{Type: DateRangeToday},
		{Type: DateRangeYesterday},
		{Type: DateRangeThisWeek},
		{Type: DateRangePrevWeek},
		{Type: DateRangeThisMonth},
		{Type: DateRangePrevMonth},
		{Type: DateRangeLast, LastDays: 7},
		{Type: DateRangeCustom, From: time.Now().AddDate(0, 0, -7), To: time.Now()},
	}

	for _, spec := range ranges {
		result, err := svc.List(spec, nil)
		if err != nil {
			t.Errorf("unexpected error for range type %d: %v", spec.Type, err)
		}
		if result == nil {
			t.Errorf("expected result for range type %d", spec.Type)
		}
	}
}

func TestEntryService_Edit(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewEntryService(storagePath, config.DefaultConfig())

	// Create an entry
	_, _ = svc.Create("original task for 1h")

	// Edit description
	e, err := svc.Edit(1, "updated task", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Description != "updated task" {
		t.Errorf("expected description 'updated task', got %q", e.Description)
	}

	// Edit duration
	e, err = svc.Edit(1, "", "2h")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.DurationMinutes != 120 {
		t.Errorf("expected 120 minutes, got %d", e.DurationMinutes)
	}

	// Edit both
	e, err = svc.Edit(1, "final task @project #tag", "30m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Description != "final task" {
		t.Errorf("expected description 'final task', got %q", e.Description)
	}
	if e.Project != "project" {
		t.Errorf("expected project 'project', got %q", e.Project)
	}
}

func TestEntryService_Edit_Errors(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewEntryService(storagePath, config.DefaultConfig())

	// Create an entry
	_, _ = svc.Create("task for 1h")

	tests := []struct {
		name        string
		index       int
		description string
		duration    string
		wantErr     error
	}{
		{
			name:        "no changes specified",
			index:       1,
			description: "",
			duration:    "",
			wantErr:     ErrNoChangesSpecified,
		},
		{
			name:        "invalid index zero",
			index:       0,
			description: "new",
			duration:    "",
			wantErr:     ErrInvalidIndex,
		},
		{
			name:        "invalid index negative",
			index:       -1,
			description: "new",
			duration:    "",
			wantErr:     ErrInvalidIndex,
		},
		{
			name:        "index out of range",
			index:       100,
			description: "new",
			duration:    "",
			wantErr:     ErrIndexOutOfRange,
		},
		{
			name:        "empty description after parse",
			index:       1,
			description: "@project #tag",
			duration:    "",
			wantErr:     ErrEmptyDescription,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Edit(tt.index, tt.description, tt.duration)
			if err == nil {
				t.Errorf("expected error %v, got nil", tt.wantErr)
			} else if !containsError(err, tt.wantErr) {
				t.Errorf("expected error containing %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestEntryService_Edit_InvalidDuration(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewEntryService(storagePath, config.DefaultConfig())

	_, _ = svc.Create("task for 1h")
	_, err := svc.Edit(1, "", "invalid")
	if err == nil {
		t.Error("expected error for invalid duration")
	}
}

func TestEntryService_Edit_NoEntries(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewEntryService(storagePath, config.DefaultConfig())

	_, err := svc.Edit(1, "new", "")
	if err == nil {
		t.Error("expected error for no entries")
	}
}

func TestEntryService_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewEntryService(storagePath, config.DefaultConfig())

	// Create entries
	_, _ = svc.Create("task 1 for 1h")
	_, _ = svc.Create("task 2 for 2h")

	// Delete first entry
	deleted, err := svc.Delete(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted.Description != "task 1" {
		t.Errorf("expected 'task 1', got %q", deleted.Description)
	}

	// List should show only one entry
	result, _ := svc.List(DateRangeSpec{Type: DateRangeToday}, nil)
	if len(result.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(result.Entries))
	}
}

func TestEntryService_Delete_Errors(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewEntryService(storagePath, config.DefaultConfig())

	// Test delete with no entries
	_, err := svc.Delete(1)
	if err == nil {
		t.Error("expected error for delete with no entries")
	}

	// Create an entry
	_, _ = svc.Create("task for 1h")

	// Test invalid indices
	_, err = svc.Delete(0)
	if err != ErrInvalidIndex {
		t.Errorf("expected ErrInvalidIndex, got %v", err)
	}

	_, err = svc.Delete(-1)
	if err != ErrInvalidIndex {
		t.Errorf("expected ErrInvalidIndex, got %v", err)
	}

	_, err = svc.Delete(100)
	if !containsError(err, ErrIndexOutOfRange) {
		t.Errorf("expected ErrIndexOutOfRange, got %v", err)
	}
}

func TestEntryService_Restore(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewEntryService(storagePath, config.DefaultConfig())

	// Create and delete an entry
	_, _ = svc.Create("task for 1h")
	_, _ = svc.Delete(1)

	// Restore it
	restored, err := svc.Restore()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if restored.Description != "task" {
		t.Errorf("expected 'task', got %q", restored.Description)
	}

	// List should show the restored entry
	result, _ := svc.List(DateRangeSpec{Type: DateRangeToday}, nil)
	if len(result.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(result.Entries))
	}
}

func TestEntryService_Restore_NoDeleted(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewEntryService(storagePath, config.DefaultConfig())

	_, err := svc.Restore()
	if err != ErrNoDeletedEntries {
		t.Errorf("expected ErrNoDeletedEntries, got %v", err)
	}
}

func TestEntryService_GetByIndex(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewEntryService(storagePath, config.DefaultConfig())

	// Create entries
	_, _ = svc.Create("task 1 for 1h")
	_, _ = svc.Create("task 2 for 2h")

	// Get by index
	ie, err := svc.GetByIndex(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ie.Entry.Description != "task 1" {
		t.Errorf("expected 'task 1', got %q", ie.Entry.Description)
	}
	if ie.ActiveIndex != 1 {
		t.Errorf("expected ActiveIndex 1, got %d", ie.ActiveIndex)
	}

	ie, err = svc.GetByIndex(2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ie.Entry.Description != "task 2" {
		t.Errorf("expected 'task 2', got %q", ie.Entry.Description)
	}
}

func TestEntryService_GetByIndex_Errors(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewEntryService(storagePath, config.DefaultConfig())

	// Invalid index
	_, err := svc.GetByIndex(0)
	if err != ErrInvalidIndex {
		t.Errorf("expected ErrInvalidIndex, got %v", err)
	}

	_, err = svc.GetByIndex(-1)
	if err != ErrInvalidIndex {
		t.Errorf("expected ErrInvalidIndex, got %v", err)
	}

	// Out of range
	_, err = svc.GetByIndex(1)
	if err != ErrIndexOutOfRange {
		t.Errorf("expected ErrIndexOutOfRange, got %v", err)
	}
}

func TestEntryService_GetActiveCount(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewEntryService(storagePath, config.DefaultConfig())

	// Initially zero
	count, err := svc.GetActiveCount()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}

	// After creating entries
	_, _ = svc.Create("task 1 for 1h")
	_, _ = svc.Create("task 2 for 2h")

	count, err = svc.GetActiveCount()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2, got %d", count)
	}
}

func TestFormatDateRangeForDisplay(t *testing.T) {
	tests := []struct {
		name  string
		start time.Time
		end   time.Time
	}{
		{
			name:  "same day",
			start: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 1, 15, 23, 59, 59, 0, time.UTC),
		},
		{
			name:  "same year",
			start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "different years",
			start: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDateRangeForDisplay(tt.start, tt.end)
			if result == "" {
				t.Error("expected non-empty result")
			}
		})
	}
}

func TestFormatDurationSimple(t *testing.T) {
	tests := []struct {
		minutes int
		want    string
	}{
		{30, "30m"},
		{60, "1h"},
		{90, "1h30m"},
		{120, "2h"},
		{150, "2h30m"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			result := formatDurationSimple(tt.minutes)
			if result != tt.want {
				t.Errorf("formatDurationSimple(%d) = %q, want %q", tt.minutes, result, tt.want)
			}
		})
	}
}

func TestEntryService_List_ReadError(t *testing.T) {
	// Use invalid path - empty file is created automatically
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewEntryService(storagePath, config.DefaultConfig())

	result, err := svc.List(DateRangeSpec{Type: DateRangeToday}, nil)
	// Should return empty result, not error (file doesn't exist yet)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result.Entries))
	}
}

func TestEntryService_buildRawInput(t *testing.T) {
	svc := NewEntryService("/tmp/test.jsonl", config.DefaultConfig())

	tests := []struct {
		name  string
		entry entry.Entry
		want  string
	}{
		{
			name: "simple entry",
			entry: entry.Entry{
				Description:     "task",
				DurationMinutes: 60,
			},
			want: "task for 1h",
		},
		{
			name: "with project",
			entry: entry.Entry{
				Description:     "task",
				DurationMinutes: 30,
				Project:         "acme",
			},
			want: "task @acme for 30m",
		},
		{
			name: "with tags",
			entry: entry.Entry{
				Description:     "task",
				DurationMinutes: 90,
				Tags:            []string{"urgent", "bug"},
			},
			want: "task #urgent #bug for 1h30m",
		},
		{
			name: "with project and tags",
			entry: entry.Entry{
				Description:     "task",
				DurationMinutes: 120,
				Project:         "acme",
				Tags:            []string{"urgent"},
			},
			want: "task @acme #urgent for 2h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.buildRawInput(tt.entry)
			if result != tt.want {
				t.Errorf("buildRawInput() = %q, want %q", result, tt.want)
			}
		})
	}
}

func TestEntryService_Create_StorageError(t *testing.T) {
	// Create a directory where the file should be to cause write error
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		t.Fatal(err)
	}

	svc := NewEntryService(storagePath, config.DefaultConfig())
	_, err := svc.Create("task for 1h")
	if err == nil {
		t.Error("expected error when storage path is a directory")
	}
}

func TestEntryService_CreateFromParts_StorageError(t *testing.T) {
	// Create a directory where the file should be to cause write error
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		t.Fatal(err)
	}

	svc := NewEntryService(storagePath, config.DefaultConfig())
	_, err := svc.CreateFromParts("task", 60, "", nil)
	if err == nil {
		t.Error("expected error when storage path is a directory")
	}
}

func TestEntryService_Edit_StorageError(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	svc := NewEntryService(storagePath, config.DefaultConfig())

	// Create an entry first
	_, _ = svc.Create("task for 1h")

	// Make directory read-only to cause write error
	if err := os.Chmod(tmpDir, 0555); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(tmpDir, 0755) }()

	_, err := svc.Edit(1, "new task", "")
	if err == nil {
		t.Error("expected error when directory is read-only")
	}
}

func TestEntryService_Delete_StorageError(t *testing.T) {
	tmpDir := t.TempDir()
	goodStoragePath := filepath.Join(tmpDir, "entries.jsonl")
	goodSvc := NewEntryService(goodStoragePath, config.DefaultConfig())

	// Create an entry first
	_, _ = goodSvc.Create("task for 1h")

	// Now create a service with a bad path (directory instead of file)
	badStoragePath := filepath.Join(tmpDir, "baddir")
	if err := os.MkdirAll(badStoragePath, 0755); err != nil {
		t.Fatal(err)
	}
	badSvc := NewEntryService(badStoragePath, config.DefaultConfig())

	_, err := badSvc.Delete(1)
	if err == nil {
		t.Error("expected error when storage path is a directory")
	}
}

func TestEntryService_GetActiveCount_StorageError(t *testing.T) {
	// Test with a directory path to cause read error
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "entries.jsonl")
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		t.Fatal(err)
	}

	svc := NewEntryService(storagePath, config.DefaultConfig())
	_, err := svc.GetActiveCount()
	if err == nil {
		t.Error("expected error when storage path is a directory")
	}
}

func containsError(err, target error) bool {
	if err == nil || target == nil {
		return err == target
	}
	return err.Error() == target.Error() ||
		(len(err.Error()) > len(target.Error()) &&
			err.Error()[len(err.Error())-len(target.Error()):] == target.Error()) ||
		(len(target.Error()) > 0 && contains(err.Error(), target.Error()))
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
