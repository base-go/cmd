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

	"github.com/spf13/cobra"
)

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

func getLatestRelease() (*Release, error) {
	url := "https://api.github.com/repos/base-go/cmd/releases/latest"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func downloadAndInstall(url, targetPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return os.Chmod(targetPath, 0755)
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade Base CLI to the latest version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Checking for updates...")

		release, err := getLatestRelease()
		if err != nil {
			fmt.Printf("Error checking for updates: %v\n", err)
			return
		}

		currentVersion := "1.0.0" // Replace with actual current version
		latestVersion := strings.TrimPrefix(release.TagName, "v")

		if currentVersion == latestVersion {
			fmt.Printf("You're already using the latest version (%s)\n", currentVersion)
			return
		}

		// Determine the correct asset name based on OS and architecture
		osName := runtime.GOOS
		archName := runtime.GOARCH
		assetPrefix := fmt.Sprintf("base_%s_%s", osName, archName)

		var downloadURL string
		for _, asset := range release.Assets {
			if strings.HasPrefix(asset.Name, assetPrefix) {
				downloadURL = asset.BrowserDownloadURL
				break
			}
		}

		if downloadURL == "" {
			fmt.Printf("No compatible binary found for your system (%s_%s)\n", osName, archName)
			return
		}

		fmt.Printf("Downloading version %s...\n", latestVersion)

		// Get the current executable path
		execPath, err := os.Executable()
		if err != nil {
			fmt.Printf("Error getting executable path: %v\n", err)
			return
		}

		// Create a temporary file for the download
		tmpDir := os.TempDir()
		tmpFile := filepath.Join(tmpDir, "base-new")

		// Download the new version
		if err := downloadAndInstall(downloadURL, tmpFile); err != nil {
			fmt.Printf("Error downloading update: %v\n", err)
			return
		}

		// Replace the old binary
		if err := os.Rename(tmpFile, execPath); err != nil {
			// If direct rename fails (e.g., on Windows), try copy and remove
			if err := copyFile(tmpFile, execPath); err != nil {
				fmt.Printf("Error installing update: %v\n", err)
				os.Remove(tmpFile)
				return
			}
			os.Remove(tmpFile)
		}

		fmt.Printf("Successfully upgraded to version %s!\n", latestVersion)
	},
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return os.Chmod(dst, 0755)
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
