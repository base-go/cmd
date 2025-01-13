package cmd

import (
	"github.com/base-go/cmd/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "base",
	Short: "Base CLI - A modern Go web framework",
	Long: `Base CLI is a powerful tool for building modern web applications in Go.
It provides scaffolding, code generation, and development utilities.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip version check for version and upgrade commands
		if cmd.Name() != "version" && cmd.Name() != "upgrade" {
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
