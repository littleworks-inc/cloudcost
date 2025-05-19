package cmd

import (
	"fmt"
	"os"
	"strings"
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
				fmt.Printf("- %s (%s): $%.4f/hour, $%.2f/month",
					resource.Name, resource.Size, resource.HourlyPrice, resource.MonthlyPrice)

				// Display pricing source or error if available
				if resource.PricingDetails != nil && strings.HasPrefix(resource.PricingDetails.PricingSource, "Error:") {
					fmt.Printf(" [%s]", resource.PricingDetails.PricingSource)
				}

				fmt.Println()
			}

			// Display warnings about missing credentials if all prices are zero
			if report.TotalMonthly == 0 && len(report.Resources) > 0 {
				fmt.Printf("\nWARNING: All resource prices are $0.00. This may be because:\n")
				fmt.Printf("- AWS credentials are not configured\n")
				fmt.Printf("- The AWS Pricing API is not accessible\n")
				fmt.Printf("- The resource types are not supported for pricing\n")

				fmt.Printf("\nTo configure AWS credentials:\n")
				fmt.Printf("1. Install the AWS CLI: https://aws.amazon.com/cli/\n")
				fmt.Printf("2. Run 'aws configure' and provide your AWS access key, secret key, and region\n")
				fmt.Printf("3. Or set the AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, and AWS_REGION environment variables\n")
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
