package cmd

import (
	"fmt"
	"strings"

	"github.com/base-go/cmd/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		info := version.GetBuildInfo()
		fmt.Println(info.String())

		// Check for updates
		release, err := version.CheckLatestVersion()
		if err != nil {
			return
		}

		latestVersion := strings.TrimPrefix(release.TagName, "v")
		if version.HasUpdate(info.Version, latestVersion) {
			fmt.Print(version.FormatUpdateMessage(
				info.Version,
				latestVersion,
				release.HTMLURL,
				release.Body,
			))
		} else {
			fmt.Printf("\n✨ You're up to date! Using the latest version %s\n", info.Version)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
