package cmd

import (
	"fmt"
	"strings"

	"github.com/BaseTechStack/basecmd/version"
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
			// Check if it's a major version upgrade
			if isMajorVersionUpgrade(info.Version, latestVersion) {
				fmt.Printf("\n🚨 MAJOR VERSION AVAILABLE: %s → %s\n", info.Version, latestVersion)
				if strings.HasPrefix(latestVersion, "2.") && strings.HasPrefix(info.Version, "1.") {
					fmt.Println("🎉 NEW in v2.0.0: Automatic Relationship Detection!")
					fmt.Println("   Fields ending with '_id' now auto-generate GORM relationships")
				}
				fmt.Println("⚠️  This is a major version with potential breaking changes.")
				fmt.Printf("📚 Changelog: %s\n", release.HTMLURL)
				fmt.Println("\nTo upgrade: base upgrade")
			} else {
				fmt.Print(version.FormatUpdateMessage(
					info.Version,
					latestVersion,
					release.HTMLURL,
					release.Body,
				))
			}
		} else {
			fmt.Printf("\n✨ You're up to date! Using the latest version %s\n", info.Version)
		}
	},
}


func init() {
	rootCmd.AddCommand(versionCmd)
}
