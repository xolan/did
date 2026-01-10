package cmd

import (
	"encoding/csv"
	"fmt"

	"github.com/xolan/did/internal/timer"
)

func writeCSVHeader(writer *csv.Writer, headers []string) error {
	if err := writer.Write(headers); err != nil {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Failed to write CSV headers")
		_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
		deps.Exit(1)
		return err
	}
	return nil
}

func writeCSVRow(writer *csv.Writer, row []string) error {
	if err := writer.Write(row); err != nil {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Failed to write CSV row")
		_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
		deps.Exit(1)
		return err
	}
	return nil
}

func loadExistingTimerState(timerPath string) (*timer.TimerState, error) {
	existingTimer, err := timer.LoadTimerState(timerPath)
	if err != nil {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Failed to load existing timer")
		_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
		deps.Exit(1)
		return nil, err
	}
	return existingTimer, nil
}

func warnClearTimerStateFailed(err error) {
	_, _ = fmt.Fprintln(deps.Stderr, "Warning: Entry saved but failed to clear timer state")
	_, _ = fmt.Fprintf(deps.Stderr, "Details: %v\n", err)
}

func handleListBackupsError(err error) {
	_, _ = fmt.Fprintf(deps.Stderr, "Error: Failed to list backups: %v\n", err)
	deps.Exit(1)
}
