package cmd

import (
	"fmt"
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
	// Get storage path
	storagePath, err := deps.StoragePath()
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: Failed to get storage path: %v\n", err)
		deps.Exit(1)
		return
	}

	// List available backups
	backups, err := storage.ListBackupsForStorage(storagePath)
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: Failed to list backups: %v\n", err)
		deps.Exit(1)
		return
	}

	if len(backups) == 0 {
		_, _ = fmt.Fprintln(deps.Stdout, "No backups available")
		deps.Exit(1)
		return
	}

	// Display available backups
	_, _ = fmt.Fprintln(deps.Stdout, "Available backups:")
	for _, backup := range backups {
		if backup.Number == 1 {
			_, _ = fmt.Fprintf(deps.Stdout, "  %d: %s (most recent)\n", backup.Number, backup.Path)
		} else {
			_, _ = fmt.Fprintf(deps.Stdout, "  %d: %s\n", backup.Number, backup.Path)
		}
	}
	_, _ = fmt.Fprintln(deps.Stdout)

	backupNum := 1
	if len(args) > 0 {
		num, err := strconv.Atoi(args[0])
		if err != nil {
			_, _ = fmt.Fprintf(deps.Stderr, "Error: Invalid backup number '%s'\n", args[0])
			deps.Exit(1)
			return
		}
		if num < 1 || num > 3 {
			_, _ = fmt.Fprintf(deps.Stderr, "Error: Backup number must be between 1 and 3 (got %d)\n", num)
			deps.Exit(1)
			return
		}
		backupNum = num
	}

	backupExists := false
	for _, backup := range backups {
		if backup.Number == backupNum {
			backupExists = true
			break
		}
	}

	if !backupExists {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: Backup %d does not exist\n", backupNum)
		deps.Exit(1)
		return
	}

	// Restore the backup
	if err := storage.RestoreBackupForStorage(storagePath, backupNum); err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "Error: Failed to restore backup: %v\n", err)
		deps.Exit(1)
		return
	}

	_, _ = fmt.Fprintf(deps.Stdout, "Successfully restored from backup %d\n", backupNum)
}
