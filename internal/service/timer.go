package service

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/xolan/did/internal/config"
	"github.com/xolan/did/internal/entry"
	"github.com/xolan/did/internal/storage"
	"github.com/xolan/did/internal/timer"
)

// Timer-specific errors
var (
	ErrTimerAlreadyRunning = errors.New("timer is already running")
	ErrNoTimerRunning      = errors.New("no timer is running")
)

// TimerService provides operations for managing the timer
type TimerService struct {
	timerPath   string
	storagePath string
	config      config.Config
}

// NewTimerService creates a new TimerService
func NewTimerService(timerPath, storagePath string, cfg config.Config) *TimerService {
	return &TimerService{
		timerPath:   timerPath,
		storagePath: storagePath,
		config:      cfg,
	}
}

// Start starts a new timer with the given description.
// If force is true, it will override any existing timer.
// Returns the existing timer state if one is running and force is false.
func (s *TimerService) Start(description string, force bool) (*timer.TimerState, *timer.TimerState, error) {
	description = trimSpace(description)
	if description == "" {
		return nil, nil, ErrEmptyDescription
	}

	// Parse project and tags from description
	cleanDesc, project, tags := entry.ParseProjectAndTags(description)
	if cleanDesc == "" {
		return nil, nil, ErrEmptyDescription
	}

	// Check if timer is already running
	isRunning, err := timer.IsTimerRunning(s.timerPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check timer status: %w", err)
	}

	var existingTimer *timer.TimerState
	if isRunning {
		existingTimer, err = timer.LoadTimerState(s.timerPath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load existing timer: %w", err)
		}

		if !force {
			return nil, existingTimer, ErrTimerAlreadyRunning
		}
	}

	// Create timer state
	state := timer.TimerState{
		StartedAt:   time.Now(),
		Description: cleanDesc,
		Project:     project,
		Tags:        tags,
	}

	// Save timer state
	if err := timer.SaveTimerState(s.timerPath, state); err != nil {
		return nil, nil, fmt.Errorf("failed to save timer state: %w", err)
	}

	return &state, existingTimer, nil
}

// Stop stops the current timer and creates a time tracking entry.
// Returns the created entry and the timer state that was stopped.
func (s *TimerService) Stop() (*entry.Entry, *timer.TimerState, error) {
	// Load timer state
	state, err := timer.LoadTimerState(s.timerPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load timer state: %w", err)
	}

	if state == nil {
		return nil, nil, ErrNoTimerRunning
	}

	// Calculate duration
	elapsed := time.Since(state.StartedAt)
	durationMinutes := calculateDurationMinutes(elapsed)

	// Create entry
	now := time.Now()
	e := entry.Entry{
		Timestamp:       now,
		Description:     state.Description,
		DurationMinutes: durationMinutes,
		RawInput:        fmt.Sprintf("%s for %s", state.Description, formatDurationSimple(durationMinutes)),
		Project:         state.Project,
		Tags:            state.Tags,
	}

	// Append entry to storage
	if err := storage.AppendEntry(s.storagePath, e); err != nil {
		return nil, nil, fmt.Errorf("failed to save entry: %w", err)
	}

	// Clear timer state - ignore error since entry was saved successfully
	// and the timer file will be overwritten on next start anyway
	_ = timer.ClearTimerState(s.timerPath)

	return &e, state, nil
}

// Cancel cancels the current timer without creating an entry
func (s *TimerService) Cancel() (*timer.TimerState, error) {
	// Load timer state first to return it
	state, err := timer.LoadTimerState(s.timerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load timer state: %w", err)
	}

	if state == nil {
		return nil, ErrNoTimerRunning
	}

	// Clear timer state
	if err := timer.ClearTimerState(s.timerPath); err != nil {
		return nil, fmt.Errorf("failed to clear timer: %w", err)
	}

	return state, nil
}

// Status returns the current timer status
func (s *TimerService) Status() (*TimerStatus, error) {
	state, err := timer.LoadTimerState(s.timerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load timer state: %w", err)
	}

	status := &TimerStatus{
		Running: state != nil,
		State:   state,
	}

	if state != nil {
		status.ElapsedTime = time.Since(state.StartedAt)
	}

	return status, nil
}

// IsRunning checks if a timer is currently running
func (s *TimerService) IsRunning() (bool, error) {
	return timer.IsTimerRunning(s.timerPath)
}

// calculateDurationMinutes calculates duration in minutes from a time.Duration.
// Rounds to the nearest minute with a minimum of 1 minute.
func calculateDurationMinutes(d time.Duration) int {
	minutes := int(math.Round(d.Minutes()))
	if minutes < 1 {
		return 1
	}
	return minutes
}

// trimSpace trims whitespace from a string (helper to avoid import in service)
func trimSpace(s string) string {
	// Simple implementation to avoid strings import
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
