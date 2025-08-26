package cmd

import (
	"fmt"
	"strings"

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
			if release, err := version.CheckLatestVersion(); err == nil {
				info := version.GetBuildInfo()
				latestVersion := strings.TrimPrefix(release.TagName, "v")
				// Only show update message if there's actually an update
				if version.HasUpdate(info.Version, latestVersion) {
					fmt.Print(version.FormatUpdateMessage(
						info.Version,
						latestVersion,
						release.HTMLURL,
						release.Body,
					))
				}
			} else {
				fmt.Println("Failed to check for updates:", err)
			}
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Commands will be added here in subsequent steps
}
