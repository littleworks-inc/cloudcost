package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/littleworks-inc/cloudcost/internal/controller"
	"github.com/littleworks-inc/cloudcost/internal/parser/terraform"
	"github.com/littleworks-inc/cloudcost/internal/pricing/aws"
	"github.com/spf13/cobra"
)

var estimatePath string
var outputFile string

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
		// Check if path exists
		if _, err := os.Stat(estimatePath); os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", estimatePath)
		}

		fmt.Printf("Estimating costs for IaC files in: %s\n", estimatePath)

		// Create estimator
		estimator := controller.NewEstimator()

		// Register parsers
		estimator.RegisterParser(terraform.NewParser())

		// Register pricing clients
		estimator.RegisterPricingClient("aws", aws.NewClient())

		// Perform estimation
		report, err := estimator.Estimate(estimatePath)
		if err != nil {
			return fmt.Errorf("estimation failed: %v", err)
		}

		// Set report timestamp
		if report != nil {
			report.Timestamp = time.Now()

			// Display summary
			fmt.Printf("\nEstimation complete:\n")
			fmt.Printf("Total Resources: %d\n", len(report.Resources))
			fmt.Printf("Estimated Monthly Cost: $%.2f\n", report.TotalMonthly)
			fmt.Printf("Estimated Yearly Cost: $%.2f\n", report.TotalYearly)

			// Display resource details
			fmt.Printf("\nResource Details:\n")
			for _, resource := range report.Resources {
				fmt.Printf("- %s (%s): $%.4f/hour, $%.2f/month\n",
					resource.Name, resource.Size, resource.HourlyPrice, resource.MonthlyPrice)
			}
		} else {
			fmt.Printf("\nNo resources found or no pricing data available.\n")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(estimateCmd)
	estimateCmd.Flags().StringVar(&estimatePath, "path", "", "Path to IaC files (required)")
	estimateCmd.Flags().StringVar(&outputFile, "output-file", "", "File to save the report to")
	estimateCmd.MarkFlagRequired("path")
}
