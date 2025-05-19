package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version information set by build flags
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Display the version, commit, and build date of cloudcost.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Cloud Cost Estimator v%s\n", Version)
		fmt.Printf("Commit: %s\n", Commit)
		fmt.Printf("Built: %s\n", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
