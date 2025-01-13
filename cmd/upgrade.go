package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
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
	ZipballURL string `json:"zipball_url"`
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade Base to the latest version",
	Long:  `Upgrade Base to the latest version from the latest GitHub release.`,
	Run:   upgradeBase,
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}

func upgradeBase(cmd *cobra.Command, args []string) {
	fmt.Println("Upgrading Base to the latest version...")
	fmt.Println("Version information:")
	fmt.Println(version.GetBuildInfo())

	// Get latest release info from GitHub
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

	// Compare versions
	currentVersion := strings.TrimPrefix(version.Version, "v")
	latestVersion := strings.TrimPrefix(release.TagName, "v")
	if currentVersion == latestVersion {
		fmt.Printf("You are already on the latest version (%s)\n", version.Version)
		return
	}

	// Get user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %v\n", err)
		return
	}

	// Define base directory
	baseDir := filepath.Join(home, ".base")

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		fmt.Printf("Error creating base directory: %v\n", err)
		return
	}

	// Define paths
	zipPath := filepath.Join(baseDir, "base.zip")
	extractPath := filepath.Join(baseDir, "cmd-main")
	binaryPath := filepath.Join(baseDir, "base")

	// Clean up existing files
	os.Remove(zipPath)
	os.RemoveAll(extractPath)
	
	// Try to remove the binary, but don't fail if we can't
	if err := os.Remove(binaryPath); err != nil {
		fmt.Println("Note: Could not remove existing binary. You may need to run with sudo.")
	}

	// Download the latest version
	fmt.Printf("Downloading version %s...\n", release.TagName)
	if err := downloadFile(zipPath, release.ZipballURL); err != nil {
		fmt.Printf("Error downloading release: %v\n", err)
		return
	}

	// Extract the zip file
	fmt.Println("Extracting the repository...")
	unzipCmd := exec.Command("unzip", "-o", zipPath, "-d", baseDir)
	unzipCmd.Stdout = os.Stdout
	unzipCmd.Stderr = os.Stderr
	if err := unzipCmd.Run(); err != nil {
		fmt.Printf("Error extracting repository: %v\n", err)
		return
	}

	// Find the extracted directory (it will have a prefix)
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		fmt.Printf("Error reading base directory: %v\n", err)
		return
	}

	var extractedDir string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "base-go-cmd-") {
			extractedDir = filepath.Join(baseDir, entry.Name())
			break
		}
	}

	if extractedDir == "" {
		fmt.Println("Error: Could not find extracted directory")
		return
	}

	// Change to the extracted directory
	if err := os.Chdir(extractedDir); err != nil {
		fmt.Printf("Error changing directory: %v\n", err)
		return
	}

	// Initialize and tidy module
	fmt.Println("Initializing and tidying module...")
	if err := exec.Command("go", "mod", "tidy").Run(); err != nil {
		fmt.Printf("Error tidying module: %v\n", err)
		return
	}

	// Build the tool
	fmt.Println("Building the tool...")
	buildCmd := exec.Command("go", "build", "-o", "base")
	buildCmd.Env = append(os.Environ(),
		"CGO_ENABLED=0",
		"GOOS="+runtime.GOOS,
		"GOARCH="+runtime.GOARCH,
	)
	if err := buildCmd.Run(); err != nil {
		fmt.Printf("Error building tool: %v\n", err)
		return
	}

	// Install the binary
	fmt.Println("Installing the tool...")
	if err := os.Rename("base", binaryPath); err != nil {
		fmt.Printf("Error installing binary: %v\n", err)
		fmt.Println("\nTo upgrade Base CLI, please run the following command in your terminal:")
		fmt.Println("  sudo base upgrade")
		fmt.Println("\nThis requires sudo privileges to replace the existing binary.")
		return
	}

	fmt.Printf("Base CLI has been successfully upgraded to version %s!\n", release.TagName)
}

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
