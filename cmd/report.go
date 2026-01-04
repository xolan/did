package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report [@project | #tag]",
	Short: "Generate detailed reports showing time spent grouped by project or tag",
	Long: `Generate detailed reports showing time spent grouped by project or tag.
Useful for client billing, team stand-ups, or personal review.

Report Types:

  Single Project/Tag Reports:
    Show all entries for a specific project or tag with totals.

    did report @project            Show all entries for a specific project
    did report #tag                Show all entries with a specific tag
    did report --project acme      Alternative syntax for project filter
    did report --tag review        Alternative syntax for tag filter

  Grouped Reports:
    Show hours grouped by all projects or tags.

    did report --by project        Show hours grouped by all projects
    did report --by tag            Show hours grouped by all tags

Date Filtering:
  Use --from and --to to filter by date range
  Use --last to filter by relative days (e.g., 'last 7 days')

Examples:

  Single Project/Tag Reports:
    did report @acme                     Show all entries for project 'acme'
    did report #review                   Show all entries tagged 'review'
    did report --project acme            Alternative syntax
    did report --tag review              Alternative syntax
    did report @acme --last 7            Project report for last 7 days
    did report #bugfix --from 2024-01-01 Tag report from specific date

  Grouped Reports:
    did report --by project              Show hours by all projects
    did report --by tag                  Show hours by all tags
    did report --by project --last 30    Project breakdown for last 30 days
    did report --by tag --from 2024-01-01 --to 2024-01-31    Tag breakdown for date range`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse shorthand filters (@project, #tag) and remove them from args
		args = parseShorthandFilters(cmd, args)
		runReport(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)

	// Add --by flag for grouping mode
	reportCmd.Flags().String("by", "", "Group by 'project' or 'tag'")

	// Note: --project and --tag flags are inherited from root command's PersistentFlags
	// Date filtering flags (--from, --to, --last) will be added in subtask 1.2
}

// runReport handles the report command logic
func runReport(cmd *cobra.Command, args []string) {
	// Get flag values
	groupBy, _ := cmd.Flags().GetString("by")
	projectFilter, _ := cmd.Root().PersistentFlags().GetString("project")
	tagFilters, _ := cmd.Root().PersistentFlags().GetStringSlice("tag")

	// Validate --by flag value if provided
	if groupBy != "" && groupBy != "project" && groupBy != "tag" {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Invalid --by value. Must be 'project' or 'tag'")
		_, _ = fmt.Fprintln(deps.Stderr, "Usage:")
		_, _ = fmt.Fprintln(deps.Stderr, "  did report --by project")
		_, _ = fmt.Fprintln(deps.Stderr, "  did report --by tag")
		deps.Exit(1)
		return
	}

	// Validate flag combinations
	if groupBy != "" && (projectFilter != "" || len(tagFilters) > 0) {
		_, _ = fmt.Fprintln(deps.Stderr, "Error: Cannot use --by with --project or --tag filters")
		_, _ = fmt.Fprintln(deps.Stderr, "Use either:")
		_, _ = fmt.Fprintln(deps.Stderr, "  did report --by project     (grouped report)")
		_, _ = fmt.Fprintln(deps.Stderr, "  did report @project         (single project report)")
		deps.Exit(1)
		return
	}

	// Determine report mode
	if groupBy == "project" {
		// Grouped by project report (subtask 3.1)
		_, _ = fmt.Fprintln(deps.Stdout, "Grouped by project report - to be implemented")
		return
	}

	if groupBy == "tag" {
		// Grouped by tag report (subtask 3.2)
		_, _ = fmt.Fprintln(deps.Stdout, "Grouped by tag report - to be implemented")
		return
	}

	if projectFilter != "" {
		// Single project report (subtask 2.1)
		_, _ = fmt.Fprintln(deps.Stdout, "Single project report - to be implemented")
		return
	}

	if len(tagFilters) > 0 {
		// Single tag report (subtask 2.2)
		_, _ = fmt.Fprintln(deps.Stdout, "Single tag report - to be implemented")
		return
	}

	// No filters provided - show usage help
	_, _ = fmt.Fprintln(deps.Stderr, "Error: No filters specified")
	_, _ = fmt.Fprintln(deps.Stderr)
	_, _ = fmt.Fprintln(deps.Stderr, "Usage:")
	_, _ = fmt.Fprintln(deps.Stderr, "  did report @project              Show all entries for a project")
	_, _ = fmt.Fprintln(deps.Stderr, "  did report #tag                  Show all entries with a tag")
	_, _ = fmt.Fprintln(deps.Stderr, "  did report --by project          Show hours grouped by all projects")
	_, _ = fmt.Fprintln(deps.Stderr, "  did report --by tag              Show hours grouped by all tags")
	_, _ = fmt.Fprintln(deps.Stderr)
	_, _ = fmt.Fprintln(deps.Stderr, "Run 'did report --help' for more information")
	deps.Exit(1)
}
