package entry

import (
	"strings"
	"testing"
)

func TestParseDuration_Hours(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"1 hour", "1h", 60},
		{"2 hours", "2h", 120},
		{"10 hours", "10h", 600},
		{"24 hours (max)", "24h", 1440},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDuration(tt.input)
			if err != nil {
				t.Errorf("ParseDuration(%q) returned unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("ParseDuration(%q) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseDuration_Minutes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"1 minute", "1m", 1},
		{"30 minutes", "30m", 30},
		{"45 minutes", "45m", 45},
		{"60 minutes", "60m", 60},
		{"90 minutes", "90m", 90},
		{"1440 minutes (24h max)", "1440m", 1440},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDuration(tt.input)
			if err != nil {
				t.Errorf("ParseDuration(%q) returned unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("ParseDuration(%q) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseDuration_CombinedFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"1h30m", "1h30m", 90},
		{"2h15m", "2h15m", 135},
		{"0h30m", "0h30m", 30},
		{"1h0m", "1h0m", 60},
		{"10h45m", "10h45m", 645},
		{"23h59m", "23h59m", 1439},
		{"24h0m", "24h0m", 1440},
		{"0h1m", "0h1m", 1},
		{"5h5m", "5h5m", 305},
		{"12h30m", "12h30m", 750},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDuration(tt.input)
			if err != nil {
				t.Errorf("ParseDuration(%q) returned unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("ParseDuration(%q) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseDuration_Invalid(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		errorSubstring string
	}{
		{"no unit", "2", "invalid time format"},
		{"invalid unit", "2x", "invalid time format"},
		{"text only", "invalid", "invalid time format"},
		{"empty string", "", "invalid time format"},
		{"just hour unit", "h", "invalid time format"},
		{"just minute unit", "m", "invalid time format"},
		{"negative hours", "-2h", "invalid time format"},
		{"decimal hours", "1.5h", "invalid time format"},
		{"space in input", "2 h", "invalid time format"},
		{"hours then text", "2hours", "invalid time format"},
		{"minutes then text", "30minutes", "invalid time format"},
		{"mixed case", "2H", "invalid time format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDuration(tt.input)
			if err == nil {
				t.Errorf("ParseDuration(%q) = %d, expected error containing %q", tt.input, result, tt.errorSubstring)
			} else if !strings.Contains(err.Error(), tt.errorSubstring) {
				t.Errorf("ParseDuration(%q) error = %q, expected to contain %q", tt.input, err.Error(), tt.errorSubstring)
			}
		})
	}
}

func TestParseDuration_Zero(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		errorSubstring string
	}{
		{"zero hours", "0h", "duration cannot be zero"},
		{"zero minutes", "0m", "duration cannot be zero"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDuration(tt.input)
			if err == nil {
				t.Errorf("ParseDuration(%q) = %d, expected error containing %q", tt.input, result, tt.errorSubstring)
			} else if !strings.Contains(err.Error(), tt.errorSubstring) {
				t.Errorf("ParseDuration(%q) error = %q, expected to contain %q", tt.input, err.Error(), tt.errorSubstring)
			}
		})
	}
}

func TestParseDuration_ExceedsMax(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		errorSubstring string
	}{
		{"25 hours", "25h", "exceeds maximum"},
		{"48 hours", "48h", "exceeds maximum"},
		{"1441 minutes", "1441m", "exceeds maximum"},
		{"2000 minutes", "2000m", "exceeds maximum"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDuration(tt.input)
			if err == nil {
				t.Errorf("ParseDuration(%q) = %d, expected error containing %q", tt.input, result, tt.errorSubstring)
			} else if !strings.Contains(err.Error(), tt.errorSubstring) {
				t.Errorf("ParseDuration(%q) error = %q, expected to contain %q", tt.input, err.Error(), tt.errorSubstring)
			}
		})
	}
}

func TestParseDuration_InvalidCombinedFormat(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		errorSubstring string
	}{
		// Exceeds max duration
		{"25h30m exceeds max", "25h30m", "exceeds maximum"},
		{"24h1m exceeds max", "24h1m", "exceeds maximum"},
		{"30h0m exceeds max", "30h0m", "exceeds maximum"},
		{"50h45m exceeds max", "50h45m", "exceeds maximum"},

		// Zero duration
		{"0h0m zero duration", "0h0m", "duration cannot be zero"},

		// Malformed patterns
		{"missing minute unit", "1h30", "invalid time format"},
		{"missing hour unit", "1 30m", "invalid time format"},
		{"wrong order", "30m1h", "invalid time format"},
		{"space between", "1h 30m", "invalid time format"},
		{"negative hours", "-1h30m", "invalid time format"},
		{"negative minutes", "1h-30m", "invalid time format"},
		{"decimal hours", "1.5h30m", "invalid time format"},
		{"decimal minutes", "1h30.5m", "invalid time format"},
		{"double hour unit", "1hh", "invalid time format"},
		{"double minute unit", "1mm", "invalid time format"},
		{"uppercase H", "1H30m", "invalid time format"},
		{"uppercase M", "1h30M", "invalid time format"},
		{"uppercase both", "1H30M", "invalid time format"},
		{"extra text after", "1h30minutes", "invalid time format"},
		{"extra text before", "time1h30m", "invalid time format"},
		{"only h and m", "hm", "invalid time format"},
		{"reversed units", "1m30h", "invalid time format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDuration(tt.input)
			if err == nil {
				t.Errorf("ParseDuration(%q) = %d, expected error containing %q", tt.input, result, tt.errorSubstring)
			} else if !strings.Contains(err.Error(), tt.errorSubstring) {
				t.Errorf("ParseDuration(%q) error = %q, expected to contain %q", tt.input, err.Error(), tt.errorSubstring)
			}
		})
	}
}

func TestMaxDurationMinutes(t *testing.T) {
	// Verify the constant is correctly set to 24 hours
	expected := 24 * 60
	if MaxDurationMinutes != expected {
		t.Errorf("MaxDurationMinutes = %d, expected %d (24 hours)", MaxDurationMinutes, expected)
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

func TestParseProjectAndTags_NoProjectNoTags(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedDesc string
		expectedProj string
		expectedTags []string
	}{
		{"simple description", "fix bug", "fix bug", "", nil},
		{"description with spaces", "working on feature implementation", "working on feature implementation", "", nil},
		{"single word", "meeting", "meeting", "", nil},
		{"empty string", "", "", "", nil},
		{"description with numbers", "fix bug 123", "fix bug 123", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc, proj, tags := ParseProjectAndTags(tt.input)
			if desc != tt.expectedDesc {
				t.Errorf("ParseProjectAndTags(%q) desc = %q, expected %q", tt.input, desc, tt.expectedDesc)
			}
			if proj != tt.expectedProj {
				t.Errorf("ParseProjectAndTags(%q) project = %q, expected %q", tt.input, proj, tt.expectedProj)
			}
			if !equalStringSlices(tags, tt.expectedTags) {
				t.Errorf("ParseProjectAndTags(%q) tags = %v, expected %v", tt.input, tags, tt.expectedTags)
			}
		})
	}
}

func TestParseProjectAndTags_ProjectOnly(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedDesc string
		expectedProj string
		expectedTags []string
	}{
		{"project at end", "fix bug @acme", "fix bug", "acme", nil},
		{"project at start", "@acme fix bug", "fix bug", "acme", nil},
		{"project in middle", "fix @acme bug", "fix bug", "acme", nil},
		{"project only", "@project", "", "project", nil},
		{"project with hyphens", "fix bug @my-project", "fix bug", "my-project", nil},
		{"project with underscores", "fix bug @my_project", "fix bug", "my_project", nil},
		{"project with numbers", "fix bug @project123", "fix bug", "project123", nil},
		{"project with mixed chars", "fix bug @my-project_v2", "fix bug", "my-project_v2", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc, proj, tags := ParseProjectAndTags(tt.input)
			if desc != tt.expectedDesc {
				t.Errorf("ParseProjectAndTags(%q) desc = %q, expected %q", tt.input, desc, tt.expectedDesc)
			}
			if proj != tt.expectedProj {
				t.Errorf("ParseProjectAndTags(%q) project = %q, expected %q", tt.input, proj, tt.expectedProj)
			}
			if !equalStringSlices(tags, tt.expectedTags) {
				t.Errorf("ParseProjectAndTags(%q) tags = %v, expected %v", tt.input, tags, tt.expectedTags)
			}
		})
	}
}

func TestParseProjectAndTags_SingleTag(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedDesc string
		expectedProj string
		expectedTags []string
	}{
		{"tag at end", "fix bug #bugfix", "fix bug", "", []string{"bugfix"}},
		{"tag at start", "#bugfix fix bug", "fix bug", "", []string{"bugfix"}},
		{"tag in middle", "fix #bugfix bug", "fix bug", "", []string{"bugfix"}},
		{"tag only", "#bugfix", "", "", []string{"bugfix"}},
		{"tag with hyphens", "fix bug #bug-fix", "fix bug", "", []string{"bug-fix"}},
		{"tag with underscores", "fix bug #bug_fix", "fix bug", "", []string{"bug_fix"}},
		{"tag with numbers", "fix bug #v2", "fix bug", "", []string{"v2"}},
		{"tag with mixed chars", "fix bug #bug-fix_v2", "fix bug", "", []string{"bug-fix_v2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc, proj, tags := ParseProjectAndTags(tt.input)
			if desc != tt.expectedDesc {
				t.Errorf("ParseProjectAndTags(%q) desc = %q, expected %q", tt.input, desc, tt.expectedDesc)
			}
			if proj != tt.expectedProj {
				t.Errorf("ParseProjectAndTags(%q) project = %q, expected %q", tt.input, proj, tt.expectedProj)
			}
			if !equalStringSlices(tags, tt.expectedTags) {
				t.Errorf("ParseProjectAndTags(%q) tags = %v, expected %v", tt.input, tags, tt.expectedTags)
			}
		})
	}
}

func TestParseProjectAndTags_MultipleTags(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedDesc string
		expectedProj string
		expectedTags []string
	}{
		{"two tags at end", "fix bug #bugfix #urgent", "fix bug", "", []string{"bugfix", "urgent"}},
		{"two tags at start", "#bugfix #urgent fix bug", "fix bug", "", []string{"bugfix", "urgent"}},
		{"tags distributed", "fix #bugfix bug #urgent", "fix bug", "", []string{"bugfix", "urgent"}},
		{"three tags", "work #feature #frontend #v2", "work", "", []string{"feature", "frontend", "v2"}},
		{"only tags", "#tag1 #tag2 #tag3", "", "", []string{"tag1", "tag2", "tag3"}},
		{"tags preserve order", "#first #second #third", "", "", []string{"first", "second", "third"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc, proj, tags := ParseProjectAndTags(tt.input)
			if desc != tt.expectedDesc {
				t.Errorf("ParseProjectAndTags(%q) desc = %q, expected %q", tt.input, desc, tt.expectedDesc)
			}
			if proj != tt.expectedProj {
				t.Errorf("ParseProjectAndTags(%q) project = %q, expected %q", tt.input, proj, tt.expectedProj)
			}
			if !equalStringSlices(tags, tt.expectedTags) {
				t.Errorf("ParseProjectAndTags(%q) tags = %v, expected %v", tt.input, tags, tt.expectedTags)
			}
		})
	}
}

func TestParseProjectAndTags_ProjectAndTags(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedDesc string
		expectedProj string
		expectedTags []string
	}{
		{"project and tag at end", "fix bug @acme #bugfix", "fix bug", "acme", []string{"bugfix"}},
		{"project and tag at start", "@acme #bugfix fix bug", "fix bug", "acme", []string{"bugfix"}},
		{"mixed order", "fix @acme bug #bugfix", "fix bug", "acme", []string{"bugfix"}},
		{"project and multiple tags", "fix bug @acme #bugfix #urgent", "fix bug", "acme", []string{"bugfix", "urgent"}},
		{"tag before project", "fix bug #bugfix @acme", "fix bug", "acme", []string{"bugfix"}},
		{"interleaved", "@acme fix #bugfix bug #urgent", "fix bug", "acme", []string{"bugfix", "urgent"}},
		{"only project and tags", "@acme #feature #urgent", "", "acme", []string{"feature", "urgent"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc, proj, tags := ParseProjectAndTags(tt.input)
			if desc != tt.expectedDesc {
				t.Errorf("ParseProjectAndTags(%q) desc = %q, expected %q", tt.input, desc, tt.expectedDesc)
			}
			if proj != tt.expectedProj {
				t.Errorf("ParseProjectAndTags(%q) project = %q, expected %q", tt.input, proj, tt.expectedProj)
			}
			if !equalStringSlices(tags, tt.expectedTags) {
				t.Errorf("ParseProjectAndTags(%q) tags = %v, expected %v", tt.input, tags, tt.expectedTags)
			}
		})
	}
}

func TestParseProjectAndTags_MultipleProjects(t *testing.T) {
	// When multiple @project tokens are found, the last one wins
	tests := []struct {
		name        string
		input       string
		expectedDesc string
		expectedProj string
		expectedTags []string
	}{
		{"two projects last wins", "fix bug @first @second", "fix bug", "second", nil},
		{"three projects last wins", "@first fix @second bug @third", "fix bug", "third", nil},
		{"projects with tags", "@first fix bug @second #urgent", "fix bug", "second", []string{"urgent"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc, proj, tags := ParseProjectAndTags(tt.input)
			if desc != tt.expectedDesc {
				t.Errorf("ParseProjectAndTags(%q) desc = %q, expected %q", tt.input, desc, tt.expectedDesc)
			}
			if proj != tt.expectedProj {
				t.Errorf("ParseProjectAndTags(%q) project = %q, expected %q", tt.input, proj, tt.expectedProj)
			}
			if !equalStringSlices(tags, tt.expectedTags) {
				t.Errorf("ParseProjectAndTags(%q) tags = %v, expected %v", tt.input, tags, tt.expectedTags)
			}
		})
	}
}

func TestParseProjectAndTags_WhitespaceHandling(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedDesc string
		expectedProj string
		expectedTags []string
	}{
		{"extra spaces between words", "fix   bug @acme", "fix bug", "acme", nil},
		{"leading spaces", "  fix bug @acme", "fix bug", "acme", nil},
		{"trailing spaces", "fix bug @acme  ", "fix bug", "acme", nil},
		{"spaces around project", "fix   @acme   bug", "fix bug", "acme", nil},
		{"spaces around tags", "fix   #tag1   bug   #tag2", "fix bug", "", []string{"tag1", "tag2"}},
		{"tabs and spaces", "fix	bug @acme", "fix bug", "acme", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc, proj, tags := ParseProjectAndTags(tt.input)
			if desc != tt.expectedDesc {
				t.Errorf("ParseProjectAndTags(%q) desc = %q, expected %q", tt.input, desc, tt.expectedDesc)
			}
			if proj != tt.expectedProj {
				t.Errorf("ParseProjectAndTags(%q) project = %q, expected %q", tt.input, proj, tt.expectedProj)
			}
			if !equalStringSlices(tags, tt.expectedTags) {
				t.Errorf("ParseProjectAndTags(%q) tags = %v, expected %v", tt.input, tags, tt.expectedTags)
			}
		})
	}
}

func TestParseProjectAndTags_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedDesc string
		expectedProj string
		expectedTags []string
	}{
		// Special characters that should NOT be part of project/tag names
		{"project with period after", "fix @acme.inc bug", "fix .inc bug", "acme", nil},
		{"tag with period after", "fix #v1.0 bug", "fix .0 bug", "", []string{"v1"}},

		// Email-like patterns (@ in email should only match valid project name part)
		{"email pattern", "sent email user@example.com", "sent email user.com", "example", nil},

		// Hash in non-tag context (only alphanumeric after # counts as tag)
		{"hash with space after", "fix # bug", "fix # bug", "", nil},
		{"at with space after", "fix @ bug", "fix @ bug", "", nil},

		// Numbers are valid tag names, so #42 is a tag
		{"number tag", "reviewed PR #42 fix", "reviewed PR fix", "", []string{"42"}},
		{"tag with number prefix", "#123 bug fix", "bug fix", "", []string{"123"}},

		// Real-world use cases
		{"version numbers", "deployed v2.0 @prod #release", "deployed v2.0", "prod", []string{"release"}},
		{"full workflow entry", "code review @client #review #urgent", "code review", "client", []string{"review", "urgent"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc, proj, tags := ParseProjectAndTags(tt.input)
			if desc != tt.expectedDesc {
				t.Errorf("ParseProjectAndTags(%q) desc = %q, expected %q", tt.input, desc, tt.expectedDesc)
			}
			if proj != tt.expectedProj {
				t.Errorf("ParseProjectAndTags(%q) project = %q, expected %q", tt.input, proj, tt.expectedProj)
			}
			if !equalStringSlices(tags, tt.expectedTags) {
				t.Errorf("ParseProjectAndTags(%q) tags = %v, expected %v", tt.input, tags, tt.expectedTags)
			}
		})
	}
}

func TestParseProjectAndTags_CaseSensitivity(t *testing.T) {
	// Project and tag names should preserve their original case
	tests := []struct {
		name        string
		input       string
		expectedDesc string
		expectedProj string
		expectedTags []string
	}{
		{"uppercase project", "fix bug @ACME", "fix bug", "ACME", nil},
		{"mixed case project", "fix bug @AcMe", "fix bug", "AcMe", nil},
		{"uppercase tag", "fix bug #URGENT", "fix bug", "", []string{"URGENT"}},
		{"mixed case tag", "fix bug #BugFix", "fix bug", "", []string{"BugFix"}},
		{"mixed case both", "fix @Client #HighPriority", "fix", "Client", []string{"HighPriority"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc, proj, tags := ParseProjectAndTags(tt.input)
			if desc != tt.expectedDesc {
				t.Errorf("ParseProjectAndTags(%q) desc = %q, expected %q", tt.input, desc, tt.expectedDesc)
			}
			if proj != tt.expectedProj {
				t.Errorf("ParseProjectAndTags(%q) project = %q, expected %q", tt.input, proj, tt.expectedProj)
			}
			if !equalStringSlices(tags, tt.expectedTags) {
				t.Errorf("ParseProjectAndTags(%q) tags = %v, expected %v", tt.input, tags, tt.expectedTags)
			}
		})
	}
}
