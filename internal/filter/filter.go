package filter

import "github.com/xolan/did/internal/entry"

// Filter represents search and filtering criteria for time tracking entries.
// All filter fields are optional - empty values match all entries.
type Filter struct {
	Keyword string   // Case-insensitive substring search in entry descriptions
	Project string   // Exact project match (case-insensitive)
	Tags    []string // All specified tags must be present (AND logic, case-insensitive)
}

// NewFilter creates a new Filter with the given criteria.
// All parameters are optional - pass empty values to match all entries.
func NewFilter(keyword, project string, tags []string) *Filter {
	return &Filter{
		Keyword: keyword,
		Project: project,
		Tags:    tags,
	}
}

// IsEmpty returns true if all filter fields are empty (matches all entries)
func (f *Filter) IsEmpty() bool {
	return f.Keyword == "" && f.Project == "" && len(f.Tags) == 0
}

// FilterEntries returns a new slice containing only entries that match the filter criteria.
// If the filter is empty, returns all entries.
func FilterEntries(entries []entry.Entry, f *Filter) []entry.Entry {
	if f.IsEmpty() {
		return entries
	}

	filtered := make([]entry.Entry, 0)
	for _, e := range entries {
		if f.Matches(e) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// Matches will be implemented in subsequent subtasks
func (f *Filter) Matches(e entry.Entry) bool {
	// Placeholder - will be implemented in subtask 1.5
	return true
}
