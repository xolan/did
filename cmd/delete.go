package cmd

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xolan/did/internal/entry"
	"github.com/xolan/did/internal/storage"
)

var yesFlag bool

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete <index>",
	Short: "Delete a time tracking entry by index",
	Long: `Delete a time tracking entry by its index number.
The index corresponds to the position in the list of entries.
A confirmation prompt will be shown unless --yes is specified.

Example:
  did delete 3
  did delete 3 --yes`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deleteEntry(args[0])
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "skip confirmation prompt")
}

// deleteEntry handles the deletion of a time tracking entry
func deleteEntry(indexStr string) {
	// Parse index from string to int
	userIndex, err := strconv.Atoi(indexStr)
	if err != nil {
		fmt.Fprintf(deps.Stderr, "Error: Invalid index '%s'. Index must be a number\n", indexStr)
		deps.Exit(1)
		return
	}

	// Validate index is positive (1-based for user)
	if userIndex < 1 {
		fmt.Fprintf(deps.Stderr, "Error: Index must be 1 or greater (got %d)\n", userIndex)
		deps.Exit(1)
		return
	}

	// Get storage path
	storagePath, err := deps.StoragePath()
	if err != nil {
		fmt.Fprintf(deps.Stderr, "Error: Failed to get storage path: %v\n", err)
		deps.Exit(1)
		return
	}

	// Read all entries to validate index bounds
	entries, err := storage.ReadEntries(storagePath)
	if err != nil {
		fmt.Fprintf(deps.Stderr, "Error: Failed to read entries: %v\n", err)
		deps.Exit(1)
		return
	}

	// Check if there are any entries
	if len(entries) == 0 {
		fmt.Fprintf(deps.Stderr, "Error: No entries to delete\n")
		deps.Exit(1)
		return
	}

	// Validate index is within bounds (convert to 0-based)
	internalIndex := userIndex - 1
	if internalIndex >= len(entries) {
		fmt.Fprintf(deps.Stderr, "Error: Index %d out of range. Valid range: 1-%d\n", userIndex, len(entries))
		deps.Exit(1)
		return
	}

	// Get the entry to delete
	entryToDelete := entries[internalIndex]

	// Show the entry being deleted
	showEntryForDeletion(entryToDelete)

	// Prompt for confirmation unless --yes flag is set
	if !yesFlag {
		if !promptConfirmation() {
			fmt.Fprintln(deps.Stdout, "Deletion cancelled")
			return
		}
	}

	// Delete the entry
	deletedEntry, err := storage.DeleteEntry(storagePath, internalIndex)
	if err != nil {
		fmt.Fprintf(deps.Stderr, "Error: Failed to delete entry: %v\n", err)
		deps.Exit(1)
		return
	}

	// Show success message
	fmt.Fprintf(deps.Stdout, "Deleted: %s (%s)\n", deletedEntry.Description, formatDuration(deletedEntry.DurationMinutes))
}

// showEntryForDeletion displays the entry that is about to be deleted
func showEntryForDeletion(e entry.Entry) {
	fmt.Fprintf(deps.Stdout, "Entry to delete:\n")
	fmt.Fprintf(deps.Stdout, "  %s  %s (%s)\n",
		e.Timestamp.Format("2006-01-02 15:04"),
		e.Description,
		formatDuration(e.DurationMinutes))
}

// promptConfirmation asks the user to confirm deletion
// Returns true if user confirms with 'y' or 'Y', false otherwise
func promptConfirmation() bool {
	fmt.Fprint(deps.Stdout, "Delete this entry? [y/N]: ")

	scanner := bufio.NewScanner(deps.Stdin)
	if !scanner.Scan() {
		return false
	}

	response := strings.TrimSpace(scanner.Text())
	return response == "y" || response == "Y"
}
