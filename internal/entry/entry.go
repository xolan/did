package entry

import "time"

// Entry represents a single time tracking entry
type Entry struct {
	Timestamp       time.Time  `json:"timestamp"`
	Description     string     `json:"description"`
	DurationMinutes int        `json:"duration_minutes"`
	RawInput        string     `json:"raw_input"`
	Project         string     `json:"project,omitempty"`
	Tags            []string   `json:"tags,omitempty"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}
