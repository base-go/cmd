package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/base-go/cmd/version"
	"github.com/spf13/cobra"
)

var Version = "dev"
var CommitHash = "none"
var BuildDate = time.Now().Format(time.RFC3339)
var GoVersion = "unknown"

type GithubRelease struct {
	TagName     string    `json:"tag_name"`
	PublishedAt time.Time `json:"published_at"`
	Body        string    `json:"body"`
	HTMLURL     string    `json:"html_url"`
}

func checkLatestVersion() (string, string, string, error) {
	url := "https://api.github.com/repos/base-go/cmd/releases/latest"
	resp, err := http.Get(url)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", err
	}

	var release GithubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return "", "", "", err
	}

	return strings.TrimPrefix(release.TagName, "v"), release.HTMLURL, release.Body, nil
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		info := version.GetBuildInfo()
		fmt.Printf("Base CLI %s\n", info.Version)
		fmt.Printf("Commit: %s\n", info.CommitHash)
		fmt.Printf("Built: %s\n", info.BuildDate)
		fmt.Printf("Go version: %s\n", info.GoVersion)

		// Check for updates
		latestVersion, releaseURL, releaseNotes, err := checkLatestVersion()
		if err != nil {
			return
		}

		if info.Version == "dev" || info.Version != latestVersion {
			fmt.Printf("\n📦 Update available! %s → %s\n", info.Version, latestVersion)
			fmt.Printf("Run: base upgrade\n")
			fmt.Printf("Release notes: %s\n", releaseURL)
			if releaseNotes != "" {
				fmt.Printf("\nChangelog:\n%s\n", releaseNotes)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
