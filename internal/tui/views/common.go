package views

import (
	"fmt"
	"strings"

	"github.com/xolan/did/internal/service"
	"github.com/xolan/did/internal/tui/ui"
)

// EntryRenderOptions configures how entries are rendered
type EntryRenderOptions struct {
	ShowDate bool // Show date in addition to time
	Width    int  // Available width for rendering
	Cursor   int  // Currently selected entry index (-1 for none)
}

// RenderEntryList renders a list of entries with aligned columns
func RenderEntryList(entries []service.IndexedEntry, styles ui.Styles, opts EntryRenderOptions) string {
	if len(entries) == 0 {
		return ""
	}

	// Calculate column widths for alignment
	maxIndexWidth := 0
	maxTimeWidth := 0
	maxDescWidth := 0

	type entryData struct {
		index    string
		time     string
		desc     string
		duration string
	}
	data := make([]entryData, len(entries))

	for i, ie := range entries {
		e := ie.Entry

		// Format index
		indexStr := fmt.Sprintf("[%d]", ie.ActiveIndex)
		if len(indexStr) > maxIndexWidth {
			maxIndexWidth = len(indexStr)
		}

		// Format time (with date if requested)
		var timeStr string
		if opts.ShowDate {
			timeStr = e.Timestamp.Format("Jan 02 15:04")
		} else {
			timeStr = e.Timestamp.Format("15:04")
		}
		if len(timeStr) > maxTimeWidth {
			maxTimeWidth = len(timeStr)
		}

		// Format description with project and tags
		descParts := []string{e.Description}
		if e.Project != "" {
			descParts = append(descParts, "@"+e.Project)
		}
		for _, tag := range e.Tags {
			descParts = append(descParts, "#"+tag)
		}
		descStr := strings.Join(descParts, " ")
		if len(descStr) > maxDescWidth {
			maxDescWidth = len(descStr)
		}

		data[i] = entryData{
			index:    indexStr,
			time:     timeStr,
			desc:     descStr,
			duration: formatDuration(e.DurationMinutes),
		}
	}

	// Limit description width to leave room for duration
	maxAllowedDescWidth := opts.Width - maxIndexWidth - maxTimeWidth - 15
	if maxAllowedDescWidth < 20 {
		maxAllowedDescWidth = 20
	}
	if maxDescWidth > maxAllowedDescWidth {
		maxDescWidth = maxAllowedDescWidth
	}

	// Render entries with alignment
	var b strings.Builder
	for i, ed := range data {
		style := styles.EntryNormal
		if i == opts.Cursor {
			style = styles.EntrySelected
		}

		// Truncate description if needed
		desc := ed.desc
		if len(desc) > maxDescWidth {
			desc = desc[:maxDescWidth-1] + "…"
		}

		// Build aligned line
		index := styles.EntryIndex.Render(fmt.Sprintf("%-*s", maxIndexWidth, ed.index))
		timeCol := styles.EntryTime.Render(fmt.Sprintf("%-*s", maxTimeWidth, ed.time))
		descCol := fmt.Sprintf("%-*s", maxDescWidth, desc)
		duration := styles.EntryDuration.Render(ed.duration)

		line := fmt.Sprintf("%s %s %s %s", index, timeCol, descCol, duration)
		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	return b.String()
}

// formatDuration formats minutes as human-readable duration
func formatDuration(minutes int) string {
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	hours := minutes / 60
	mins := minutes % 60
	if mins == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh %dm", hours, mins)
}

func pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	return word + "s"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
