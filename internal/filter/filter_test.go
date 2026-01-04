package filter

import (
	"testing"
	"time"

	"github.com/xolan/did/internal/entry"
)

// Helper function to create test entries
func makeEntry(desc, project string, tags []string) entry.Entry {
	return entry.Entry{
		Timestamp:       time.Now(),
		Description:     desc,
		DurationMinutes: 60,
		RawInput:        desc,
		Project:         project,
		Tags:            tags,
	}
}

// Helper function to compare string slices
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestNewFilter(t *testing.T) {
	tests := []struct {
		name    string
		keyword string
		project string
		tags    []string
	}{
		{
			name:    "empty filter",
			keyword: "",
			project: "",
			tags:    nil,
		},
		{
			name:    "keyword only",
			keyword: "meeting",
			project: "",
			tags:    nil,
		},
		{
			name:    "project only",
			keyword: "",
			project: "acme",
			tags:    nil,
		},
		{
			name:    "tags only",
			keyword: "",
			project: "",
			tags:    []string{"urgent", "bugfix"},
		},
		{
			name:    "all fields",
			keyword: "meeting",
			project: "acme",
			tags:    []string{"urgent"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFilter(tt.keyword, tt.project, tt.tags)
			if f.Keyword != tt.keyword {
				t.Errorf("NewFilter() keyword = %q, expected %q", f.Keyword, tt.keyword)
			}
			if f.Project != tt.project {
				t.Errorf("NewFilter() project = %q, expected %q", f.Project, tt.project)
			}
			if !equalStringSlices(f.Tags, tt.tags) {
				t.Errorf("NewFilter() tags = %v, expected %v", f.Tags, tt.tags)
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		filter   *Filter
		expected bool
	}{
		{
			name:     "empty filter",
			filter:   NewFilter("", "", nil),
			expected: true,
		},
		{
			name:     "empty filter with empty slice",
			filter:   NewFilter("", "", []string{}),
			expected: true,
		},
		{
			name:     "keyword only",
			filter:   NewFilter("meeting", "", nil),
			expected: false,
		},
		{
			name:     "project only",
			filter:   NewFilter("", "acme", nil),
			expected: false,
		},
		{
			name:     "tags only",
			filter:   NewFilter("", "", []string{"urgent"}),
			expected: false,
		},
		{
			name:     "all fields populated",
			filter:   NewFilter("meeting", "acme", []string{"urgent"}),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.filter.IsEmpty()
			if result != tt.expected {
				t.Errorf("IsEmpty() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestMatchesKeyword_EmptyKeyword(t *testing.T) {
	f := NewFilter("", "", nil)
	entries := []entry.Entry{
		makeEntry("fix bug", "", nil),
		makeEntry("meeting with client", "", nil),
		makeEntry("", "", nil),
	}

	for _, e := range entries {
		t.Run("empty keyword matches "+e.Description, func(t *testing.T) {
			if !f.MatchesKeyword(e) {
				t.Errorf("MatchesKeyword() = false, expected true for empty keyword")
			}
		})
	}
}

func TestMatchesKeyword_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name        string
		keyword     string
		description string
		expected    bool
	}{
		{
			name:        "exact lowercase match",
			keyword:     "meeting",
			description: "meeting with client",
			expected:    true,
		},
		{
			name:        "exact uppercase match",
			keyword:     "MEETING",
			description: "meeting with client",
			expected:    true,
		},
		{
			name:        "mixed case keyword",
			keyword:     "MeEtInG",
			description: "meeting with client",
			expected:    true,
		},
		{
			name:        "mixed case description",
			keyword:     "meeting",
			description: "MEETING with client",
			expected:    true,
		},
		{
			name:        "both mixed case",
			keyword:     "BuG",
			description: "Fix BuG in parser",
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFilter(tt.keyword, "", nil)
			e := makeEntry(tt.description, "", nil)
			result := f.MatchesKeyword(e)
			if result != tt.expected {
				t.Errorf("MatchesKeyword() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestMatchesKeyword_SubstringMatch(t *testing.T) {
	tests := []struct {
		name        string
		keyword     string
		description string
		expected    bool
	}{
		{
			name:        "keyword at start",
			keyword:     "fix",
			description: "fix bug in parser",
			expected:    true,
		},
		{
			name:        "keyword at end",
			keyword:     "parser",
			description: "fix bug in parser",
			expected:    true,
		},
		{
			name:        "keyword in middle",
			keyword:     "bug",
			description: "fix bug in parser",
			expected:    true,
		},
		{
			name:        "keyword is full description",
			keyword:     "meeting",
			description: "meeting",
			expected:    true,
		},
		{
			name:        "partial word match",
			keyword:     "meet",
			description: "meeting with client",
			expected:    true,
		},
		{
			name:        "no match",
			keyword:     "testing",
			description: "fix bug in parser",
			expected:    false,
		},
		{
			name:        "empty description no match",
			keyword:     "meeting",
			description: "",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFilter(tt.keyword, "", nil)
			e := makeEntry(tt.description, "", nil)
			result := f.MatchesKeyword(e)
			if result != tt.expected {
				t.Errorf("MatchesKeyword() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestMatchesProject_EmptyProject(t *testing.T) {
	f := NewFilter("", "", nil)
	entries := []entry.Entry{
		makeEntry("fix bug", "acme", nil),
		makeEntry("meeting", "clientA", nil),
		makeEntry("work", "", nil),
	}

	for _, e := range entries {
		t.Run("empty project matches "+e.Project, func(t *testing.T) {
			if !f.MatchesProject(e) {
				t.Errorf("MatchesProject() = false, expected true for empty project filter")
			}
		})
	}
}

func TestMatchesProject_ExactMatch(t *testing.T) {
	tests := []struct {
		name            string
		filterProject   string
		entryProject    string
		expected        bool
		expectedMessage string
	}{
		{
			name:          "exact match",
			filterProject: "acme",
			entryProject:  "acme",
			expected:      true,
		},
		{
			name:          "no match",
			filterProject: "acme",
			entryProject:  "clientB",
			expected:      false,
		},
		{
			name:          "substring not matched",
			filterProject: "acme",
			entryProject:  "acme-corp",
			expected:      false,
		},
		{
			name:          "superstring not matched",
			filterProject: "acme-corp",
			entryProject:  "acme",
			expected:      false,
		},
		{
			name:          "empty entry project",
			filterProject: "acme",
			entryProject:  "",
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFilter("", tt.filterProject, nil)
			e := makeEntry("work", tt.entryProject, nil)
			result := f.MatchesProject(e)
			if result != tt.expected {
				t.Errorf("MatchesProject() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestMatchesProject_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name          string
		filterProject string
		entryProject  string
		expected      bool
	}{
		{
			name:          "exact lowercase match",
			filterProject: "acme",
			entryProject:  "acme",
			expected:      true,
		},
		{
			name:          "exact uppercase match",
			filterProject: "ACME",
			entryProject:  "acme",
			expected:      true,
		},
		{
			name:          "mixed case filter",
			filterProject: "AcMe",
			entryProject:  "acme",
			expected:      true,
		},
		{
			name:          "mixed case entry",
			filterProject: "acme",
			entryProject:  "ACME",
			expected:      true,
		},
		{
			name:          "both mixed case",
			filterProject: "AcMe",
			entryProject:  "aCmE",
			expected:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFilter("", tt.filterProject, nil)
			e := makeEntry("work", tt.entryProject, nil)
			result := f.MatchesProject(e)
			if result != tt.expected {
				t.Errorf("MatchesProject() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestMatchesTags_EmptyTags(t *testing.T) {
	f := NewFilter("", "", nil)
	entries := []entry.Entry{
		makeEntry("fix bug", "", []string{"urgent"}),
		makeEntry("meeting", "", []string{"meeting", "client"}),
		makeEntry("work", "", nil),
	}

	for _, e := range entries {
		t.Run("empty tags matches entry with tags", func(t *testing.T) {
			if !f.MatchesTags(e) {
				t.Errorf("MatchesTags() = false, expected true for empty tags filter")
			}
		})
	}
}

func TestMatchesTags_SingleTag(t *testing.T) {
	tests := []struct {
		name      string
		filterTag string
		entryTags []string
		expected  bool
	}{
		{
			name:      "tag present",
			filterTag: "urgent",
			entryTags: []string{"urgent"},
			expected:  true,
		},
		{
			name:      "tag present with others",
			filterTag: "urgent",
			entryTags: []string{"urgent", "bugfix"},
			expected:  true,
		},
		{
			name:      "tag not present",
			filterTag: "urgent",
			entryTags: []string{"feature"},
			expected:  false,
		},
		{
			name:      "entry has no tags",
			filterTag: "urgent",
			entryTags: nil,
			expected:  false,
		},
		{
			name:      "entry has empty tags",
			filterTag: "urgent",
			entryTags: []string{},
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFilter("", "", []string{tt.filterTag})
			e := makeEntry("work", "", tt.entryTags)
			result := f.MatchesTags(e)
			if result != tt.expected {
				t.Errorf("MatchesTags() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestMatchesTags_MultipleTags_ANDLogic(t *testing.T) {
	tests := []struct {
		name       string
		filterTags []string
		entryTags  []string
		expected   bool
	}{
		{
			name:       "all tags present",
			filterTags: []string{"urgent", "bugfix"},
			entryTags:  []string{"urgent", "bugfix"},
			expected:   true,
		},
		{
			name:       "all tags present plus extra",
			filterTags: []string{"urgent", "bugfix"},
			entryTags:  []string{"urgent", "bugfix", "feature"},
			expected:   true,
		},
		{
			name:       "all tags present different order",
			filterTags: []string{"urgent", "bugfix"},
			entryTags:  []string{"bugfix", "urgent"},
			expected:   true,
		},
		{
			name:       "only first tag present",
			filterTags: []string{"urgent", "bugfix"},
			entryTags:  []string{"urgent"},
			expected:   false,
		},
		{
			name:       "only second tag present",
			filterTags: []string{"urgent", "bugfix"},
			entryTags:  []string{"bugfix"},
			expected:   false,
		},
		{
			name:       "no tags present",
			filterTags: []string{"urgent", "bugfix"},
			entryTags:  []string{"feature"},
			expected:   false,
		},
		{
			name:       "entry has no tags",
			filterTags: []string{"urgent", "bugfix"},
			entryTags:  nil,
			expected:   false,
		},
		{
			name:       "three tags all present",
			filterTags: []string{"urgent", "bugfix", "backend"},
			entryTags:  []string{"urgent", "bugfix", "backend"},
			expected:   true,
		},
		{
			name:       "three tags only two present",
			filterTags: []string{"urgent", "bugfix", "backend"},
			entryTags:  []string{"urgent", "bugfix"},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFilter("", "", tt.filterTags)
			e := makeEntry("work", "", tt.entryTags)
			result := f.MatchesTags(e)
			if result != tt.expected {
				t.Errorf("MatchesTags() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestMatchesTags_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name       string
		filterTags []string
		entryTags  []string
		expected   bool
	}{
		{
			name:       "exact lowercase match",
			filterTags: []string{"urgent"},
			entryTags:  []string{"urgent"},
			expected:   true,
		},
		{
			name:       "filter uppercase entry lowercase",
			filterTags: []string{"URGENT"},
			entryTags:  []string{"urgent"},
			expected:   true,
		},
		{
			name:       "filter lowercase entry uppercase",
			filterTags: []string{"urgent"},
			entryTags:  []string{"URGENT"},
			expected:   true,
		},
		{
			name:       "both mixed case",
			filterTags: []string{"UrGeNt"},
			entryTags:  []string{"uRgEnT"},
			expected:   true,
		},
		{
			name:       "multiple tags mixed case",
			filterTags: []string{"URGENT", "bugfix"},
			entryTags:  []string{"urgent", "BUGFIX"},
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFilter("", "", tt.filterTags)
			e := makeEntry("work", "", tt.entryTags)
			result := f.MatchesTags(e)
			if result != tt.expected {
				t.Errorf("MatchesTags() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestMatches_EmptyFilter(t *testing.T) {
	f := NewFilter("", "", nil)
	entries := []entry.Entry{
		makeEntry("fix bug", "acme", []string{"urgent"}),
		makeEntry("meeting", "clientA", []string{"meeting"}),
		makeEntry("work", "", nil),
		makeEntry("", "", []string{}),
	}

	for i, e := range entries {
		t.Run("empty filter matches all entries", func(t *testing.T) {
			if !f.Matches(e) {
				t.Errorf("Matches() = false for entry %d, expected true for empty filter", i)
			}
		})
	}
}

func TestMatches_KeywordOnly(t *testing.T) {
	f := NewFilter("meeting", "", nil)
	tests := []struct {
		name     string
		entry    entry.Entry
		expected bool
	}{
		{
			name:     "keyword matches",
			entry:    makeEntry("meeting with client", "acme", []string{"urgent"}),
			expected: true,
		},
		{
			name:     "keyword matches no project no tags",
			entry:    makeEntry("meeting", "", nil),
			expected: true,
		},
		{
			name:     "keyword does not match",
			entry:    makeEntry("fix bug", "acme", []string{"urgent"}),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.Matches(tt.entry)
			if result != tt.expected {
				t.Errorf("Matches() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestMatches_ProjectOnly(t *testing.T) {
	f := NewFilter("", "acme", nil)
	tests := []struct {
		name     string
		entry    entry.Entry
		expected bool
	}{
		{
			name:     "project matches",
			entry:    makeEntry("fix bug", "acme", []string{"urgent"}),
			expected: true,
		},
		{
			name:     "project matches no tags",
			entry:    makeEntry("work", "acme", nil),
			expected: true,
		},
		{
			name:     "project does not match",
			entry:    makeEntry("fix bug", "clientB", []string{"urgent"}),
			expected: false,
		},
		{
			name:     "entry has no project",
			entry:    makeEntry("work", "", []string{"urgent"}),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.Matches(tt.entry)
			if result != tt.expected {
				t.Errorf("Matches() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestMatches_TagsOnly(t *testing.T) {
	f := NewFilter("", "", []string{"urgent"})
	tests := []struct {
		name     string
		entry    entry.Entry
		expected bool
	}{
		{
			name:     "tag matches",
			entry:    makeEntry("fix bug", "acme", []string{"urgent"}),
			expected: true,
		},
		{
			name:     "tag matches no project",
			entry:    makeEntry("work", "", []string{"urgent", "bugfix"}),
			expected: true,
		},
		{
			name:     "tag does not match",
			entry:    makeEntry("fix bug", "acme", []string{"feature"}),
			expected: false,
		},
		{
			name:     "entry has no tags",
			entry:    makeEntry("work", "acme", nil),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.Matches(tt.entry)
			if result != tt.expected {
				t.Errorf("Matches() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestMatches_KeywordAndProject(t *testing.T) {
	f := NewFilter("meeting", "acme", nil)
	tests := []struct {
		name     string
		entry    entry.Entry
		expected bool
	}{
		{
			name:     "both match",
			entry:    makeEntry("meeting with client", "acme", []string{"urgent"}),
			expected: true,
		},
		{
			name:     "keyword matches project does not",
			entry:    makeEntry("meeting with client", "clientB", []string{"urgent"}),
			expected: false,
		},
		{
			name:     "project matches keyword does not",
			entry:    makeEntry("fix bug", "acme", []string{"urgent"}),
			expected: false,
		},
		{
			name:     "neither matches",
			entry:    makeEntry("fix bug", "clientB", []string{"urgent"}),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.Matches(tt.entry)
			if result != tt.expected {
				t.Errorf("Matches() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestMatches_KeywordAndTags(t *testing.T) {
	f := NewFilter("meeting", "", []string{"urgent"})
	tests := []struct {
		name     string
		entry    entry.Entry
		expected bool
	}{
		{
			name:     "both match",
			entry:    makeEntry("meeting with client", "acme", []string{"urgent"}),
			expected: true,
		},
		{
			name:     "keyword matches tags do not",
			entry:    makeEntry("meeting with client", "acme", []string{"feature"}),
			expected: false,
		},
		{
			name:     "tags match keyword does not",
			entry:    makeEntry("fix bug", "acme", []string{"urgent"}),
			expected: false,
		},
		{
			name:     "neither matches",
			entry:    makeEntry("fix bug", "acme", []string{"feature"}),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.Matches(tt.entry)
			if result != tt.expected {
				t.Errorf("Matches() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestMatches_ProjectAndTags(t *testing.T) {
	f := NewFilter("", "acme", []string{"urgent"})
	tests := []struct {
		name     string
		entry    entry.Entry
		expected bool
	}{
		{
			name:     "both match",
			entry:    makeEntry("fix bug", "acme", []string{"urgent"}),
			expected: true,
		},
		{
			name:     "project matches tags do not",
			entry:    makeEntry("fix bug", "acme", []string{"feature"}),
			expected: false,
		},
		{
			name:     "tags match project does not",
			entry:    makeEntry("fix bug", "clientB", []string{"urgent"}),
			expected: false,
		},
		{
			name:     "neither matches",
			entry:    makeEntry("fix bug", "clientB", []string{"feature"}),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.Matches(tt.entry)
			if result != tt.expected {
				t.Errorf("Matches() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestMatches_AllFilters(t *testing.T) {
	f := NewFilter("meeting", "acme", []string{"urgent"})
	tests := []struct {
		name     string
		entry    entry.Entry
		expected bool
	}{
		{
			name:     "all match",
			entry:    makeEntry("meeting with client", "acme", []string{"urgent"}),
			expected: true,
		},
		{
			name:     "all match with extra tags",
			entry:    makeEntry("meeting with client", "acme", []string{"urgent", "client"}),
			expected: true,
		},
		{
			name:     "keyword does not match",
			entry:    makeEntry("fix bug", "acme", []string{"urgent"}),
			expected: false,
		},
		{
			name:     "project does not match",
			entry:    makeEntry("meeting with client", "clientB", []string{"urgent"}),
			expected: false,
		},
		{
			name:     "tags do not match",
			entry:    makeEntry("meeting with client", "acme", []string{"feature"}),
			expected: false,
		},
		{
			name:     "only keyword matches",
			entry:    makeEntry("meeting", "clientB", []string{"feature"}),
			expected: false,
		},
		{
			name:     "only project matches",
			entry:    makeEntry("fix bug", "acme", []string{"feature"}),
			expected: false,
		},
		{
			name:     "only tags match",
			entry:    makeEntry("fix bug", "clientB", []string{"urgent"}),
			expected: false,
		},
		{
			name:     "none match",
			entry:    makeEntry("fix bug", "clientB", []string{"feature"}),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.Matches(tt.entry)
			if result != tt.expected {
				t.Errorf("Matches() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestMatches_ComplexScenarios(t *testing.T) {
	tests := []struct {
		name     string
		filter   *Filter
		entry    entry.Entry
		expected bool
	}{
		{
			name:     "multiple tags all present with keyword and project",
			filter:   NewFilter("code", "acme", []string{"urgent", "bugfix"}),
			entry:    makeEntry("code review", "acme", []string{"urgent", "bugfix", "backend"}),
			expected: true,
		},
		{
			name:     "multiple tags one missing",
			filter:   NewFilter("code", "acme", []string{"urgent", "bugfix"}),
			entry:    makeEntry("code review", "acme", []string{"urgent"}),
			expected: false,
		},
		{
			name:     "case insensitive all filters",
			filter:   NewFilter("MEETING", "ACME", []string{"URGENT"}),
			entry:    makeEntry("meeting notes", "acme", []string{"urgent"}),
			expected: true,
		},
		{
			name:     "partial keyword match with filters",
			filter:   NewFilter("bug", "acme", []string{"urgent"}),
			entry:    makeEntry("fix bugfix in parser", "acme", []string{"urgent"}),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.filter.Matches(tt.entry)
			if result != tt.expected {
				t.Errorf("Matches() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestFilterEntries_EmptyFilter(t *testing.T) {
	f := NewFilter("", "", nil)
	entries := []entry.Entry{
		makeEntry("fix bug", "acme", []string{"urgent"}),
		makeEntry("meeting", "clientA", []string{"meeting"}),
		makeEntry("work", "", nil),
	}

	result := FilterEntries(entries, f)
	if len(result) != len(entries) {
		t.Errorf("FilterEntries() returned %d entries, expected %d", len(result), len(entries))
	}

	// Verify all entries are returned
	for i := range entries {
		if result[i].Description != entries[i].Description {
			t.Errorf("FilterEntries() entry %d mismatch", i)
		}
	}
}

func TestFilterEntries_FilterByKeyword(t *testing.T) {
	f := NewFilter("meeting", "", nil)
	entries := []entry.Entry{
		makeEntry("fix bug", "acme", []string{"urgent"}),
		makeEntry("meeting with client", "clientA", []string{"meeting"}),
		makeEntry("team meeting", "", nil),
		makeEntry("work on feature", "acme", []string{"feature"}),
	}

	result := FilterEntries(entries, f)
	if len(result) != 2 {
		t.Errorf("FilterEntries() returned %d entries, expected 2", len(result))
	}

	// Verify correct entries are returned
	if result[0].Description != "meeting with client" {
		t.Errorf("FilterEntries() first result = %q, expected 'meeting with client'", result[0].Description)
	}
	if result[1].Description != "team meeting" {
		t.Errorf("FilterEntries() second result = %q, expected 'team meeting'", result[1].Description)
	}
}

func TestFilterEntries_FilterByProject(t *testing.T) {
	f := NewFilter("", "acme", nil)
	entries := []entry.Entry{
		makeEntry("fix bug", "acme", []string{"urgent"}),
		makeEntry("meeting", "clientA", []string{"meeting"}),
		makeEntry("work", "acme", nil),
		makeEntry("task", "clientB", []string{"feature"}),
	}

	result := FilterEntries(entries, f)
	if len(result) != 2 {
		t.Errorf("FilterEntries() returned %d entries, expected 2", len(result))
	}

	// Verify all results have project "acme"
	for i, e := range result {
		if e.Project != "acme" {
			t.Errorf("FilterEntries() result %d project = %q, expected 'acme'", i, e.Project)
		}
	}
}

func TestFilterEntries_FilterByTags(t *testing.T) {
	f := NewFilter("", "", []string{"urgent"})
	entries := []entry.Entry{
		makeEntry("fix bug", "acme", []string{"urgent"}),
		makeEntry("meeting", "clientA", []string{"meeting"}),
		makeEntry("critical fix", "acme", []string{"urgent", "bugfix"}),
		makeEntry("task", "clientB", []string{"feature"}),
	}

	result := FilterEntries(entries, f)
	if len(result) != 2 {
		t.Errorf("FilterEntries() returned %d entries, expected 2", len(result))
	}

	// Verify all results have "urgent" tag
	for i, e := range result {
		hasUrgent := false
		for _, tag := range e.Tags {
			if tag == "urgent" {
				hasUrgent = true
				break
			}
		}
		if !hasUrgent {
			t.Errorf("FilterEntries() result %d missing 'urgent' tag", i)
		}
	}
}

func TestFilterEntries_CombinedFilters(t *testing.T) {
	f := NewFilter("fix", "acme", []string{"urgent"})
	entries := []entry.Entry{
		makeEntry("fix bug", "acme", []string{"urgent"}),
		makeEntry("fix issue", "clientA", []string{"urgent"}),
		makeEntry("fix problem", "acme", []string{"feature"}),
		makeEntry("meeting", "acme", []string{"urgent"}),
		makeEntry("quick fix", "acme", []string{"urgent", "hotfix"}),
	}

	result := FilterEntries(entries, f)
	if len(result) != 2 {
		t.Errorf("FilterEntries() returned %d entries, expected 2", len(result))
	}

	// Verify results match all criteria
	expectedDescriptions := []string{"fix bug", "quick fix"}
	for i, e := range result {
		if e.Description != expectedDescriptions[i] {
			t.Errorf("FilterEntries() result %d = %q, expected %q", i, e.Description, expectedDescriptions[i])
		}
		if e.Project != "acme" {
			t.Errorf("FilterEntries() result %d project = %q, expected 'acme'", i, e.Project)
		}
		hasUrgent := false
		for _, tag := range e.Tags {
			if tag == "urgent" {
				hasUrgent = true
				break
			}
		}
		if !hasUrgent {
			t.Errorf("FilterEntries() result %d missing 'urgent' tag", i)
		}
	}
}

func TestFilterEntries_NoMatches(t *testing.T) {
	f := NewFilter("meeting", "acme", []string{"urgent"})
	entries := []entry.Entry{
		makeEntry("fix bug", "acme", []string{"urgent"}),
		makeEntry("meeting", "clientA", []string{"urgent"}),
		makeEntry("meeting", "acme", []string{"feature"}),
	}

	result := FilterEntries(entries, f)
	if len(result) != 0 {
		t.Errorf("FilterEntries() returned %d entries, expected 0", len(result))
	}
}

func TestFilterEntries_EmptySlice(t *testing.T) {
	f := NewFilter("meeting", "", nil)
	entries := []entry.Entry{}

	result := FilterEntries(entries, f)
	if len(result) != 0 {
		t.Errorf("FilterEntries() returned %d entries, expected 0", len(result))
	}
}

func TestFilterEntries_NilSlice(t *testing.T) {
	f := NewFilter("meeting", "", nil)
	var entries []entry.Entry

	result := FilterEntries(entries, f)
	if result == nil {
		t.Error("FilterEntries() returned nil, expected empty slice")
	}
	if len(result) != 0 {
		t.Errorf("FilterEntries() returned %d entries, expected 0", len(result))
	}
}
