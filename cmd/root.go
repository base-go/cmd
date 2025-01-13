package cmd

import (
	"base/version"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "base [command] [args]",
	Short: "Generate or destroy modules for the application",
	Long:  `A command-line tool to generate new modules with predefined structure or destroy existing modules for the application.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip version check for version command
		if cmd.Name() != "version" {
			version.PrintUpdateMessage()
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Commands will be added here in subsequent steps
}
