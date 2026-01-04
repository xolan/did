package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xolan/did/internal/storage"
)

// undoCmd represents the undo command
var undoCmd = &cobra.Command{
	Use:   "undo",
	Short: "Restore the most recently deleted entry",
	Long: `Restore the most recently deleted entry.
This command recovers the last entry that was deleted using 'did delete'.

Example:
  did undo`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		undoDelete()
	},
}

// undoDelete restores the most recently soft-deleted entry
func undoDelete() {
	// Get storage path
	storagePath, err := deps.StoragePath()
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: Failed to get storage path: %v\n", err)
		deps.Exit(1)
		return
	}

	// Find the most recently deleted entry
	_, index, err := storage.GetMostRecentlyDeleted(storagePath)
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: %v\n", err)
		_, _ = fmt.Fprintln(deps.Stderr, "Hint: No entries to restore. Delete an entry first with 'did delete <index>'")
		deps.Exit(1)
		return
	}

	// Restore the entry
	restoredEntry, err := storage.RestoreEntry(storagePath, index)
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: Failed to restore entry: %v\n", err)
		deps.Exit(1)
		return
	}

	// Show success message with entry details
	_, _ = fmt.Fprintf(deps.Stdout, "Restored: %s (%s)\n",
		formatEntryForLog(restoredEntry.Description, restoredEntry.Project, restoredEntry.Tags),
		formatDuration(restoredEntry.DurationMinutes))
	_, _ = fmt.Fprintf(deps.Stdout, "  Timestamp: %s\n", restoredEntry.Timestamp.Format("2006-01-02 15:04"))
}
