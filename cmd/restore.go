package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/xolan/did/internal/storage"
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore [backup_number]",
	Short: "Restore from a backup file",
	Long: `Restore the storage file from a backup.

By default, restores from the most recent backup (.bak.1).
Optionally specify a backup number to restore from (1-3).

Examples:
  did restore       Restore from most recent backup
  did restore 2     Restore from backup #2`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		restoreFromBackup(args)
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}

// restoreFromBackup handles the restore command logic
func restoreFromBackup(args []string) {
	// List available backups
	backups, err := storage.ListBackups()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to list backups: %v\n", err)
		os.Exit(1)
	}

	if len(backups) == 0 {
		fmt.Println("No backups available")
		os.Exit(1)
	}

	// Display available backups
	fmt.Println("Available backups:")
	for _, backup := range backups {
		if backup.Number == 1 {
			fmt.Printf("  %d: %s (most recent)\n", backup.Number, backup.Path)
		} else {
			fmt.Printf("  %d: %s\n", backup.Number, backup.Path)
		}
	}
	fmt.Println()

	// Determine which backup to restore
	backupNum := 1 // Default to most recent
	if len(args) > 0 {
		num, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Invalid backup number '%s'\n", args[0])
			os.Exit(1)
		}
		backupNum = num
	}

	// Validate backup exists
	backupExists := false
	for _, backup := range backups {
		if backup.Number == backupNum {
			backupExists = true
			break
		}
	}

	if !backupExists {
		fmt.Fprintf(os.Stderr, "Error: Backup %d does not exist\n", backupNum)
		os.Exit(1)
	}

	// Restore the backup
	if err := storage.RestoreBackup(backupNum); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to restore backup: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully restored from backup %d\n", backupNum)
}
