package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/base-go/cmd/version"
	"github.com/spf13/cobra"
)

type GithubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade Base CLI to the latest version",
	Long:  `Upgrade Base CLI to the latest version from GitHub releases.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("\nUpgrading Base to the latest version...")
		fmt.Println("Version information:")
		info := version.GetBuildInfo()
		fmt.Println(info)

		// Get the latest release from GitHub
		resp, err := http.Get("https://api.github.com/repos/base-go/cmd/releases/latest")
		if err != nil {
			fmt.Printf("Error checking latest version: %v\n", err)
			return
		}
		defer resp.Body.Close()

		var release GithubRelease
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			fmt.Printf("Error parsing release info: %v\n", err)
			return
		}

		if release.TagName == "" {
			fmt.Println("No release found")
			return
		}

		fmt.Printf("Downloading version %s...\n", release.TagName)

		// Find the correct asset for the current platform
		var downloadURL string
		platform := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
		for _, asset := range release.Assets {
			if strings.Contains(asset.Name, platform) {
				downloadURL = asset.BrowserDownloadURL
				break
			}
		}

		if downloadURL == "" {
			fmt.Printf("No binary found for platform %s\n", platform)
			return
		}

		// Download the binary
		resp, err = http.Get(downloadURL)
		if err != nil {
			fmt.Printf("Error downloading release: %v\n", err)
			return
		}
		defer resp.Body.Close()

		// Create a temporary file
		tmpFile, err := os.CreateTemp("", "base-*")
		if err != nil {
			fmt.Printf("Error creating temporary file: %v\n", err)
			return
		}
		defer os.Remove(tmpFile.Name())

		// Copy the downloaded binary to the temporary file
		_, err = io.Copy(tmpFile, resp.Body)
		if err != nil {
			fmt.Printf("Error saving binary: %v\n", err)
			return
		}
		tmpFile.Close()

		// Make the temporary file executable
		err = os.Chmod(tmpFile.Name(), 0755)
		if err != nil {
			fmt.Printf("Error making binary executable: %v\n", err)
			return
		}

		// Get the path to the current binary
		currentBinary, err := os.Executable()
		if err != nil {
			fmt.Printf("Error getting current binary path: %v\n", err)
			return
		}
		currentBinary, err = filepath.EvalSymlinks(currentBinary)
		if err != nil {
			fmt.Printf("Error resolving symlinks: %v\n", err)
			return
		}

		// Replace the current binary
		if err := os.Rename(tmpFile.Name(), currentBinary); err != nil {
			if os.IsPermission(err) {
				fmt.Println("\nTo upgrade Base CLI, please run the following command in your terminal:")
				fmt.Println("  sudo base upgrade")
				fmt.Println("\nThis requires sudo privileges to replace the existing binary.")
			} else {
				fmt.Printf("Error installing binary: %v\n", err)
			}
			return
		}

		fmt.Printf("Base CLI has been successfully upgraded to version %s!\n", release.TagName)
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
