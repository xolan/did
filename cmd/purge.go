package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xolan/did/internal/storage"
)

var purgeYesFlag bool

// purgeCmd represents the purge command
var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Permanently remove all soft-deleted entries",
	Long: `Permanently remove all soft-deleted entries from storage.
This action cannot be undone. A confirmation prompt will be shown
unless --yes is specified.

Deleted entries are normally kept for 7 days to allow recovery with
'did undo'. Use this command when you want to immediately and
permanently remove all deleted entries.

Example:
  did purge
  did purge --yes`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		purgeDeleted()
	},
}

func init() {
	purgeCmd.Flags().BoolVarP(&purgeYesFlag, "yes", "y", false, "skip confirmation prompt")
}

// purgeDeleted permanently removes all soft-deleted entries
func purgeDeleted() {
	// Get storage path
	storagePath, err := deps.StoragePath()
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: Failed to get storage path: %v\n", err)
		deps.Exit(1)
		return
	}

	// Prompt for confirmation unless --yes flag is set
	if !purgeYesFlag {
		if !promptPurgeConfirmation() {
			_, _ = fmt.Fprintln(deps.Stdout, "Purge cancelled")
			return
		}
	}

	// Purge all soft-deleted entries
	count, err := storage.PurgeDeletedEntries(storagePath)
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: Failed to purge entries: %v\n", err)
		deps.Exit(1)
		return
	}

	// Show success message with count
	if count == 0 {
		_, _ = fmt.Fprintln(deps.Stdout, "No deleted entries to purge")
	} else if count == 1 {
		_, _ = fmt.Fprintln(deps.Stdout, "Purged 1 entry")
	} else {
		_, _ = fmt.Fprintf(deps.Stdout, "Purged %d entries\n", count)
	}
}

// promptPurgeConfirmation asks the user to confirm purge operation
// Returns true if user confirms with 'y' or 'Y', false otherwise
func promptPurgeConfirmation() bool {
	_, _ = fmt.Fprint(deps.Stdout, "Permanently delete all soft-deleted entries? This cannot be undone. [y/N]: ")

	scanner := bufio.NewScanner(deps.Stdin)
	if !scanner.Scan() {
		return false
	}

	response := strings.TrimSpace(scanner.Text())
	return response == "y" || response == "Y"
}
