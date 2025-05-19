package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var estimatePath string

// estimateCmd represents the estimate command
var estimateCmd = &cobra.Command{
	Use:   "estimate",
	Short: "Estimate cloud costs from IaC files",
	Long: `Estimate cloud costs by parsing Infrastructure-as-Code files and 
using real-time pricing data from cloud provider APIs.

Examples:
  cloudcost estimate --path ./terraform-project
  cloudcost estimate --path ./ansible-playbooks --output json
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Estimating costs for IaC files in: %s\n", estimatePath)

		// This is where we'll add the logic to:
		// 1. Detect the IaC format
		// 2. Parse the files
		// 3. Extract resources
		// 4. Call pricing APIs
		// 5. Calculate costs
		// 6. Generate report

		return nil
	},
}

func init() {
	rootCmd.AddCommand(estimateCmd)
	estimateCmd.Flags().StringVar(&estimatePath, "path", "", "Path to IaC files (required)")
	estimateCmd.MarkFlagRequired("path")
}
