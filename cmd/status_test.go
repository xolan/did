package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/xolan/did/internal/timer"
)

func TestShowStatus_TimerRunning(t *testing.T) {
	cleanup := setupTimerTest(t)
	defer cleanup()

	// Create a timer that started 90 minutes ago
	timerPath, _ := timer.GetTimerPath()
	startTime := time.Now().Add(-90 * time.Minute)
	state := timer.TimerState{
		StartedAt:   startTime,
		Description: "fixing authentication bug",
		Project:     "",
		Tags:        []string{},
	}
	if err := timer.SaveTimerState(timerPath, state); err != nil {
		t.Fatalf("Failed to create timer: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
		TimerPath: timer.GetTimerPath,
		Config:    DefaultDeps().Config,
	}
	SetDeps(d)
	defer ResetDeps()

	showStatus()

	output := stdout.String()
	if !strings.Contains(output, "Timer running:") {
		t.Errorf("Expected 'Timer running:' in output, got: %s", output)
	}
	if !strings.Contains(output, "fixing authentication bug") {
		t.Errorf("Expected description in output, got: %s", output)
	}
	if !strings.Contains(output, "Started:") {
		t.Errorf("Expected 'Started:' in output, got: %s", output)
	}
	if !strings.Contains(output, "Elapsed:") {
		t.Errorf("Expected 'Elapsed:' in output, got: %s", output)
	}
	if !strings.Contains(output, "1h 30m") {
		t.Errorf("Expected elapsed time '1h 30m' in output, got: %s", output)
	}
}

func TestShowStatus_ElapsedTimeFormatting(t *testing.T) {
	tests := []struct {
		name            string
		elapsed         time.Duration
		expectedContain string
	}{
		{"5 minutes", 5 * time.Minute, "5m"},
		{"30 minutes", 30 * time.Minute, "30m"},
		{"1 hour", 60 * time.Minute, "1h"},
		{"1 hour 30 minutes", 90 * time.Minute, "1h 30m"},
		{"2 hours 15 minutes", 135 * time.Minute, "2h 15m"},
		{"5 hours", 5 * time.Hour, "5h"},
		{"10 hours 45 minutes", 10*time.Hour + 45*time.Minute, "10h 45m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTimerTest(t)
			defer cleanup()

			// Create a timer with specific elapsed time
			timerPath, _ := timer.GetTimerPath()
			startTime := time.Now().Add(-tt.elapsed)
			state := timer.TimerState{
				StartedAt:   startTime,
				Description: "test task",
				Project:     "",
				Tags:        []string{},
			}
			if err := timer.SaveTimerState(timerPath, state); err != nil {
				t.Fatalf("Failed to create timer: %v", err)
			}

			stdout := &bytes.Buffer{}
			d := &Deps{
				Stdout: stdout,
				Stderr: &bytes.Buffer{},
				Stdin:  strings.NewReader(""),
				Exit:   func(code int) {},
				StoragePath: func() (string, error) {
					return "", nil
				},
				TimerPath: timer.GetTimerPath,
				Config:    DefaultDeps().Config,
			}
			SetDeps(d)
			defer ResetDeps()

			showStatus()

			output := stdout.String()
			if !strings.Contains(output, tt.expectedContain) {
				t.Errorf("Expected elapsed time '%s' in output, got: %s", tt.expectedContain, output)
			}
		})
	}
}

func TestShowStatus_WithProjectAndTags(t *testing.T) {
	cleanup := setupTimerTest(t)
	defer cleanup()

	// Create a timer with project and tags
	timerPath, _ := timer.GetTimerPath()
	startTime := time.Now().Add(-2 * time.Hour)
	state := timer.TimerState{
		StartedAt:   startTime,
		Description: "code review",
		Project:     "acme",
		Tags:        []string{"review", "urgent"},
	}
	if err := timer.SaveTimerState(timerPath, state); err != nil {
		t.Fatalf("Failed to create timer: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
		TimerPath: timer.GetTimerPath,
		Config:    DefaultDeps().Config,
	}
	SetDeps(d)
	defer ResetDeps()

	showStatus()

	output := stdout.String()
	if !strings.Contains(output, "Timer running:") {
		t.Errorf("Expected 'Timer running:' in output, got: %s", output)
	}
	if !strings.Contains(output, "code review") {
		t.Errorf("Expected description in output, got: %s", output)
	}
	if !strings.Contains(output, "@acme") {
		t.Errorf("Expected @acme in output, got: %s", output)
	}
	if !strings.Contains(output, "#review") {
		t.Errorf("Expected #review in output, got: %s", output)
	}
	if !strings.Contains(output, "#urgent") {
		t.Errorf("Expected #urgent in output, got: %s", output)
	}
	if !strings.Contains(output, "2h") {
		t.Errorf("Expected elapsed time '2h' in output, got: %s", output)
	}
}

func TestShowStatus_NoTimerRunning(t *testing.T) {
	cleanup := setupTimerTest(t)
	defer cleanup()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
		TimerPath: timer.GetTimerPath,
		Config:    DefaultDeps().Config,
	}
	SetDeps(d)
	defer ResetDeps()

	showStatus()

	output := stdout.String()
	if !strings.Contains(output, "No timer running") {
		t.Errorf("Expected 'No timer running' in output, got: %s", output)
	}
	if !strings.Contains(output, "did start") {
		t.Errorf("Expected hint about 'did start' in output, got: %s", output)
	}

	// Should not contain timer details
	if strings.Contains(output, "Timer running:") {
		t.Errorf("Should not show 'Timer running:' when no timer exists, got: %s", output)
	}
	if strings.Contains(output, "Started:") {
		t.Errorf("Should not show 'Started:' when no timer exists, got: %s", output)
	}
	if strings.Contains(output, "Elapsed:") {
		t.Errorf("Should not show 'Elapsed:' when no timer exists, got: %s", output)
	}
}

func TestShowStatus_StartTimeFormatting_Today(t *testing.T) {
	cleanup := setupTimerTest(t)
	defer cleanup()

	// Create a timer that started today (use fixed time to avoid midnight edge cases)
	timerPath, _ := timer.GetTimerPath()
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())
	startTime := startOfDay.Add(-2 * time.Hour) // 10 AM today
	state := timer.TimerState{
		StartedAt:   startTime,
		Description: "today's task",
		Project:     "",
		Tags:        []string{},
	}
	if err := timer.SaveTimerState(timerPath, state); err != nil {
		t.Fatalf("Failed to create timer: %v", err)
	}

	stdout := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
		TimerPath: timer.GetTimerPath,
		Config:    DefaultDeps().Config,
	}
	SetDeps(d)
	defer ResetDeps()

	showStatus()

	output := stdout.String()
	if !strings.Contains(output, "today at") {
		t.Errorf("Expected 'today at' for timer started today, got: %s", output)
	}
	// Should contain time in format like "3:04 PM"
	if !strings.Contains(output, "M") {
		t.Errorf("Expected time with AM/PM in output, got: %s", output)
	}
}

func TestShowStatus_StartTimeFormatting_PastDate(t *testing.T) {
	cleanup := setupTimerTest(t)
	defer cleanup()

	// Create a timer that started yesterday
	timerPath, _ := timer.GetTimerPath()
	yesterday := time.Now().AddDate(0, 0, -1)
	state := timer.TimerState{
		StartedAt:   yesterday,
		Description: "yesterday's task",
		Project:     "",
		Tags:        []string{},
	}
	if err := timer.SaveTimerState(timerPath, state); err != nil {
		t.Fatalf("Failed to create timer: %v", err)
	}

	stdout := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
		TimerPath: timer.GetTimerPath,
		Config:    DefaultDeps().Config,
	}
	SetDeps(d)
	defer ResetDeps()

	showStatus()

	output := stdout.String()
	// Should NOT contain "today at" for yesterday's timer
	if strings.Contains(output, "today at") {
		t.Errorf("Should not show 'today at' for timer started yesterday, got: %s", output)
	}
	// Should contain day name and month (e.g., "Mon Jan 2 at 3:04 PM")
	expectedMonth := yesterday.Format("Jan")
	if !strings.Contains(output, expectedMonth) {
		t.Errorf("Expected month '%s' in output for past date, got: %s", expectedMonth, output)
	}
}

func TestShowStatus_SpecialCharacters(t *testing.T) {
	cleanup := setupTimerTest(t)
	defer cleanup()

	// Create a timer with special characters in description
	timerPath, _ := timer.GetTimerPath()
	state := timer.TimerState{
		StartedAt:   time.Now().Add(-30 * time.Minute),
		Description: "fixing bug with special chars: <>&\"'",
		Project:     "test-project",
		Tags:        []string{"bug-fix", "high-priority"},
	}
	if err := timer.SaveTimerState(timerPath, state); err != nil {
		t.Fatalf("Failed to create timer: %v", err)
	}

	stdout := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
		TimerPath: timer.GetTimerPath,
		Config:    DefaultDeps().Config,
	}
	SetDeps(d)
	defer ResetDeps()

	showStatus()

	output := stdout.String()
	if !strings.Contains(output, "fixing bug with special chars: <>&\"'") {
		t.Errorf("Expected special characters in description to be preserved, got: %s", output)
	}
	if !strings.Contains(output, "@test-project") {
		t.Errorf("Expected project with hyphen to be shown, got: %s", output)
	}
	if !strings.Contains(output, "#bug-fix") {
		t.Errorf("Expected tag with hyphen to be shown, got: %s", output)
	}
	if !strings.Contains(output, "#high-priority") {
		t.Errorf("Expected tag with hyphen to be shown, got: %s", output)
	}
}

func TestStatusCommand_Run(t *testing.T) {
	cleanup := setupTimerTest(t)
	defer cleanup()

	// Create a timer
	timerPath, _ := timer.GetTimerPath()
	state := timer.TimerState{
		StartedAt:   time.Now().Add(-45 * time.Minute),
		Description: "test task",
		Project:     "",
		Tags:        []string{},
	}
	if err := timer.SaveTimerState(timerPath, state); err != nil {
		t.Fatalf("Failed to create timer: %v", err)
	}

	stdout := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
		TimerPath: timer.GetTimerPath,
		Config:    DefaultDeps().Config,
	}
	SetDeps(d)
	defer ResetDeps()

	// Call the status command's Run function directly
	statusCmd.Run(statusCmd, []string{})

	if !strings.Contains(stdout.String(), "Timer running:") {
		t.Errorf("Expected 'Timer running:', got: %s", stdout.String())
	}
}

func TestShowStatus_ZeroElapsedTime(t *testing.T) {
	cleanup := setupTimerTest(t)
	defer cleanup()

	// Create a timer that just started (0 elapsed time)
	timerPath, _ := timer.GetTimerPath()
	state := timer.TimerState{
		StartedAt:   time.Now(),
		Description: "just started",
		Project:     "",
		Tags:        []string{},
	}
	if err := timer.SaveTimerState(timerPath, state); err != nil {
		t.Fatalf("Failed to create timer: %v", err)
	}

	stdout := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
		TimerPath: timer.GetTimerPath,
		Config:    DefaultDeps().Config,
	}
	SetDeps(d)
	defer ResetDeps()

	showStatus()

	output := stdout.String()
	if !strings.Contains(output, "Timer running:") {
		t.Errorf("Expected 'Timer running:' in output, got: %s", output)
	}
	// Should show 0m for just started timer
	if !strings.Contains(output, "0m") {
		t.Errorf("Expected '0m' for just started timer, got: %s", output)
	}
}

func TestShowStatus_MultipleTagsOrdering(t *testing.T) {
	cleanup := setupTimerTest(t)
	defer cleanup()

	// Create a timer with multiple tags
	timerPath, _ := timer.GetTimerPath()
	state := timer.TimerState{
		StartedAt:   time.Now().Add(-1 * time.Hour),
		Description: "multi-tag task",
		Project:     "myproject",
		Tags:        []string{"feature", "backend", "api", "urgent"},
	}
	if err := timer.SaveTimerState(timerPath, state); err != nil {
		t.Fatalf("Failed to create timer: %v", err)
	}

	stdout := &bytes.Buffer{}
	d := &Deps{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) {},
		StoragePath: func() (string, error) {
			return "", nil
		},
		TimerPath: timer.GetTimerPath,
		Config:    DefaultDeps().Config,
	}
	SetDeps(d)
	defer ResetDeps()

	showStatus()

	output := stdout.String()
	// Verify all tags are present
	if !strings.Contains(output, "#feature") {
		t.Errorf("Expected #feature in output, got: %s", output)
	}
	if !strings.Contains(output, "#backend") {
		t.Errorf("Expected #backend in output, got: %s", output)
	}
	if !strings.Contains(output, "#api") {
		t.Errorf("Expected #api in output, got: %s", output)
	}
	if !strings.Contains(output, "#urgent") {
		t.Errorf("Expected #urgent in output, got: %s", output)
	}
	if !strings.Contains(output, "@myproject") {
		t.Errorf("Expected @myproject in output, got: %s", output)
	}
}

func TestShowStatus_TimerPathError(t *testing.T) {
	exitCalled := false
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: &bytes.Buffer{},
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) { exitCalled = true },
		StoragePath: func() (string, error) {
			return "", nil
		},
		TimerPath: func() (string, error) {
			return "", os.ErrPermission
		},
		Config: DefaultDeps().Config,
	}
	SetDeps(d)
	defer ResetDeps()

	showStatus()

	if !exitCalled {
		t.Error("Expected exit to be called when TimerPath fails")
	}
	errOutput := stderr.String()
	if !strings.Contains(errOutput, "Failed to determine timer location") {
		t.Errorf("Expected timer location error, got: %s", errOutput)
	}
}

func TestShowStatus_LoadTimerStateError(t *testing.T) {
	cleanup := setupTimerTest(t)
	defer cleanup()

	// Create a corrupted timer file
	timerPath, _ := timer.GetTimerPath()
	_ = os.WriteFile(timerPath, []byte("not valid json"), 0644)

	exitCalled := false
	stderr := &bytes.Buffer{}
	d := &Deps{
		Stdout: &bytes.Buffer{},
		Stderr: stderr,
		Stdin:  strings.NewReader(""),
		Exit:   func(code int) { exitCalled = true },
		StoragePath: func() (string, error) {
			return "", nil
		},
		TimerPath: timer.GetTimerPath,
		Config:    DefaultDeps().Config,
	}
	SetDeps(d)
	defer ResetDeps()

	showStatus()

	if !exitCalled {
		t.Error("Expected exit to be called when LoadTimerState fails")
	}
	errOutput := stderr.String()
	if !strings.Contains(errOutput, "Failed to load timer state") {
		t.Errorf("Expected load timer state error, got: %s", errOutput)
	}
}
