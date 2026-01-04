package entry

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// combinedTimePattern matches combined time duration in XhYm format (e.g., "1h30m", "2h15m")
var combinedTimePattern = regexp.MustCompile(`^(\d+)h(\d+)m$`)

// timePattern matches time duration in Yh (hours) or Ym (minutes) format
var timePattern = regexp.MustCompile(`^(\d+)(h|m)$`)

// MaxDurationMinutes is the maximum allowed duration per entry (24 hours)
const MaxDurationMinutes = 24 * 60

// ParseDuration parses a time duration string in Yh, Ym, or XhYm format
// and returns the duration in minutes.
// Valid inputs: "2h" (returns 120), "30m" (returns 30), "1h30m" (returns 90)
// Invalid inputs: "invalid", "0h", "0m", "0h0m", values exceeding 24h
func ParseDuration(input string) (minutes int, err error) {
	// First try combined pattern (e.g., "1h30m")
	combinedMatches := combinedTimePattern.FindStringSubmatch(input)
	if combinedMatches != nil {
		hours, err := strconv.Atoi(combinedMatches[1])
		if err != nil {
			return 0, fmt.Errorf("invalid time format: expected Xh, Xm, or XhYm, got %s", input)
		}

		mins, err := strconv.Atoi(combinedMatches[2])
		if err != nil {
			return 0, fmt.Errorf("invalid time format: expected Xh, Xm, or XhYm, got %s", input)
		}

		// Calculate total minutes
		minutes = hours*60 + mins

		if minutes == 0 {
			return 0, fmt.Errorf("invalid duration: duration cannot be zero")
		}

		if minutes > MaxDurationMinutes {
			return 0, fmt.Errorf("invalid duration: exceeds maximum of 24 hours (%d minutes)", MaxDurationMinutes)
		}

		return minutes, nil
	}

	// Fall back to simple pattern (e.g., "2h" or "30m")
	matches := timePattern.FindStringSubmatch(input)
	if matches == nil {
		return 0, fmt.Errorf("invalid time format: expected Xh, Xm, or XhYm, got %s", input)
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("invalid time format: expected Xh, Xm, or XhYm, got %s", input)
	}

	unit := matches[2]
	if unit == "h" {
		minutes = value * 60
	} else {
		minutes = value
	}

	if minutes == 0 {
		return 0, fmt.Errorf("invalid duration: duration cannot be zero")
	}

	if minutes > MaxDurationMinutes {
		return 0, fmt.Errorf("invalid duration: exceeds maximum of 24 hours (%d minutes)", MaxDurationMinutes)
	}

	return minutes, nil
}

// projectPattern matches @project syntax (e.g., "@acme", "@my-project", "@project123")
// Project names can contain alphanumeric characters, hyphens, and underscores
var projectPattern = regexp.MustCompile(`@([a-zA-Z0-9_-]+)`)

// tagPattern matches #tag syntax (e.g., "#bugfix", "#urgent", "#v1-release")
// Tag names can contain alphanumeric characters, hyphens, and underscores
var tagPattern = regexp.MustCompile(`#([a-zA-Z0-9_-]+)`)

// ParseProjectAndTags extracts @project and #tags from a description string.
// Returns the cleaned description (without @project and #tags), the project name (if any),
// and a slice of tags.
// If multiple @project tokens are found, the last one wins.
// Example: "fix bug @acme #bugfix #urgent" -> ("fix bug", "acme", ["bugfix", "urgent"])
func ParseProjectAndTags(description string) (cleanDesc string, project string, tags []string) {
	// Extract all projects (last one wins)
	projectMatches := projectPattern.FindAllStringSubmatch(description, -1)
	if len(projectMatches) > 0 {
		project = projectMatches[len(projectMatches)-1][1]
	}

	// Extract all tags
	tagMatches := tagPattern.FindAllStringSubmatch(description, -1)
	for _, match := range tagMatches {
		tags = append(tags, match[1])
	}

	// Remove all @project and #tag tokens from the description
	cleanDesc = projectPattern.ReplaceAllString(description, "")
	cleanDesc = tagPattern.ReplaceAllString(cleanDesc, "")

	// Clean up excess whitespace
	cleanDesc = strings.TrimSpace(cleanDesc)
	// Replace multiple spaces with a single space
	cleanDesc = regexp.MustCompile(`\s+`).ReplaceAllString(cleanDesc, " ")

	return cleanDesc, project, tags
}
