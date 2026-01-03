package cmd

import (
	"bufio"
	"fmt"
	"os"
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
		fmt.Fprintf(os.Stderr, "Error: Invalid index '%s'. Index must be a number\n", indexStr)
		os.Exit(1)
	}

	// Validate index is positive (1-based for user)
	if userIndex < 1 {
		fmt.Fprintf(os.Stderr, "Error: Index must be 1 or greater (got %d)\n", userIndex)
		os.Exit(1)
	}

	// Get storage path
	storagePath, err := storage.GetStoragePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get storage path: %v\n", err)
		os.Exit(1)
	}

	// Read all entries to validate index bounds
	entries, err := storage.ReadEntries(storagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to read entries: %v\n", err)
		os.Exit(1)
	}

	// Check if there are any entries
	if len(entries) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No entries to delete\n")
		os.Exit(1)
	}

	// Validate index is within bounds (convert to 0-based)
	internalIndex := userIndex - 1
	if internalIndex >= len(entries) {
		fmt.Fprintf(os.Stderr, "Error: Index %d out of range. Valid range: 1-%d\n", userIndex, len(entries))
		os.Exit(1)
	}

	// Get the entry to delete
	entryToDelete := entries[internalIndex]

	// Show the entry being deleted
	showEntryForDeletion(entryToDelete)

	// Prompt for confirmation unless --yes flag is set
	if !yesFlag {
		if !promptConfirmation() {
			fmt.Println("Deletion cancelled")
			return
		}
	}

	// Delete the entry
	deletedEntry, err := storage.DeleteEntry(storagePath, internalIndex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to delete entry: %v\n", err)
		os.Exit(1)
	}

	// Show success message
	fmt.Printf("Deleted: %s (%s)\n", deletedEntry.Description, formatDuration(deletedEntry.DurationMinutes))
}

// showEntryForDeletion displays the entry that is about to be deleted
func showEntryForDeletion(e entry.Entry) {
	fmt.Printf("Entry to delete:\n")
	fmt.Printf("  %s  %s (%s)\n",
		e.Timestamp.Format("2006-01-02 15:04"),
		e.Description,
		formatDuration(e.DurationMinutes))
}

// promptConfirmation asks the user to confirm deletion
// Returns true if user confirms with 'y' or 'Y', false otherwise
func promptConfirmation() bool {
	fmt.Print("Delete this entry? [y/N]: ")

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return false
	}

	response := strings.TrimSpace(scanner.Text())
	return response == "y" || response == "Y"
}
