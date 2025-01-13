package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/base-go/cmd/version"
	"github.com/spf13/cobra"
)

type releaseInfo struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Description string `json:"body"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Base CLI",
	Long:  `All software has versions. This is Base's.`,
	Run: func(cmd *cobra.Command, args []string) {
		info := version.GetBuildInfo()
		fmt.Println(info)

		// Get release notes from GitHub
		resp, err := http.Get("https://api.github.com/repos/base-go/cmd/releases/tags/" + info.Version)
		if err == nil {
			defer resp.Body.Close()
			var release releaseInfo
			if err := json.NewDecoder(resp.Body).Decode(&release); err == nil && release.Description != "" {
				fmt.Println("\nWhat's new in this version:")
				notes := strings.Split(release.Description, "\n")
				for _, line := range notes {
					if strings.TrimSpace(line) != "" {
						fmt.Println(line)
					}
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
