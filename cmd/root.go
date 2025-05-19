package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var outputFormat string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cloudcost",
	Short: "Estimate cloud costs from IaC files",
	Long: `Cloud Cost Estimator for Infrastructure-as-Code (IaC)

Provides accurate, real-time cost estimates using live pricing data 
from AWS, Azure, and GCP APIs based on your IaC files like:
- Terraform
- Pulumi
- CloudFormation
- Azure Bicep
- Ansible
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cloudcost.yaml)")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "output", "text", "Output format (text, json, csv, html)")

	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cloudcost" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".cloudcost")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
