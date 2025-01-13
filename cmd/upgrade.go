package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/base-go/cmd/version"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade Base to the latest version",
	Long:  `Upgrade Base to the latest version by re-running the installation script.`,
	Run:   upgradeBase,
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}

func upgradeBase(cmd *cobra.Command, args []string) {
	fmt.Println("Upgrading Base to the latest version...")
	fmt.Println("Version information:")
	fmt.Println(version.GetBuildInfo())

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
	fmt.Println("Downloading the repository...")
	downloadURL := "https://github.com/base-go/cmd/archive/refs/heads/main.zip"
	downloadCmd := exec.Command("curl", "-L", downloadURL, "-o", zipPath)
	downloadCmd.Stdout = os.Stdout
	downloadCmd.Stderr = os.Stderr
	if err := downloadCmd.Run(); err != nil {
		fmt.Printf("Error downloading repository: %v\n", err)
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

	// Change to the extracted directory
	if err := os.Chdir(extractPath); err != nil {
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
		fmt.Println("You may need to run with sudo privileges:")
		fmt.Println("sudo base upgrade")
		return
	}

	fmt.Println("Base CLI has been successfully upgraded!")
}
