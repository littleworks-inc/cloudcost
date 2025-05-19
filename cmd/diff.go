package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var diffPath string
var compareTo string

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Compare costs between IaC versions",
	Long: `Compare estimated costs between current IaC files and a previous cost report.

Examples:
  cloudcost diff --path ./terraform-project --compare-to previous-report.json
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Comparing costs for %s against %s\n", diffPath, compareTo)

		// This is where we'll add the logic to:
		// 1. Parse current IaC files
		// 2. Load previous report
		// 3. Compare resources and costs
		// 4. Generate diff report

		return nil
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
	diffCmd.Flags().StringVar(&diffPath, "path", "", "Path to IaC files (required)")
	diffCmd.Flags().StringVar(&compareTo, "compare-to", "", "Previous cost report to compare against (required)")
	diffCmd.MarkFlagRequired("path")
	diffCmd.MarkFlagRequired("compare-to")
}
