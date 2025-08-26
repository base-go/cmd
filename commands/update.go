package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/base-go/cmd/utils"
	"github.com/base-go/cmd/version"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Base Core to the latest version",
	Long:  `Update Base Core to the latest version. This command will update the core directory of your Base project to the latest version available on GitHub.`,
	Run:   updateBaseCore,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func updateBaseCore(cmd *cobra.Command, args []string) {
	fmt.Println("Updating Base Core...")
	err := updateCore()
	if err != nil {
		fmt.Printf("Error updating Base Core: %v\n", err)
		return
	}
	fmt.Println("Base Core updated successfully.")
}

func updateCore() error {
	// Determine framework tag from CLI version (same behavior as `base new`)
	rawVersion := version.Version
	normalized := strings.TrimPrefix(rawVersion, "v")
	tag := "v" + normalized
	archiveURL := fmt.Sprintf("https://github.com/base-go/base-core/archive/refs/tags/%s.zip", tag)

	// Create a temporary working directory
	tempDir, err := os.MkdirTemp("", "base-core-update-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Download the tagged archive
	fmt.Printf("Downloading core from: %s\n", archiveURL)
	resp, err := http.Get(archiveURL)
	if err != nil {
		return fmt.Errorf("failed to download core archive: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d downloading core archive from %s", resp.StatusCode, archiveURL)
	}

	// Save to a temporary zip file
	tmpZip, err := os.CreateTemp("", "base-core-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp zip file: %v", err)
	}
	_, err = io.Copy(tmpZip, resp.Body)
	if cerr := tmpZip.Close(); cerr != nil && err == nil {
		err = cerr
	}
	if err != nil {
		return fmt.Errorf("failed to save core archive: %v", err)
	}
	defer os.Remove(tmpZip.Name())

	// Extract into tempDir
	if err := utils.Unzip(tmpZip.Name(), tempDir); err != nil {
		return fmt.Errorf("failed to extract core archive: %v", err)
	}

	// Locate extracted root directory (supports base-vX.Y.Z or base-X.Y.Z)
	candidates := []string{
		filepath.Join(tempDir, fmt.Sprintf("base-%s", tag)),        // base-v2.1.7
		filepath.Join(tempDir, fmt.Sprintf("base-%s", normalized)), // base-2.1.7
	}
	var extractedDir string
	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && fi.IsDir() {
			extractedDir = c
			break
		}
	}
	if extractedDir == "" {
		if entries, err := os.ReadDir(tempDir); err == nil {
			for _, e := range entries {
				if e.IsDir() && strings.HasPrefix(e.Name(), "base-") {
					extractedDir = filepath.Join(tempDir, e.Name())
					break
				}
			}
		}
	}
	if extractedDir == "" {
		return fmt.Errorf("could not locate extracted base directory for tag %s", tag)
	}

	// Source core directory inside extracted archive
	srcCoreDir := filepath.Join(extractedDir, "core")
	if fi, err := os.Stat(srcCoreDir); err != nil || !fi.IsDir() {
		return fmt.Errorf("core directory not found in archive at %s", srcCoreDir)
	}

	// Path to the project's core directory
	projectCoreDir := filepath.Join(".", "core")

	// Backup existing core (if present)
	backupDir := projectCoreDir + ".bak"
	if _, err := os.Stat(projectCoreDir); err == nil {
		if err := os.Rename(projectCoreDir, backupDir); err != nil {
			return fmt.Errorf("failed to backup current core directory: %v", err)
		}
	}

	// Copy new core into project
	copyErr := filepath.Walk(srcCoreDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcCoreDir, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(projectCoreDir, rel)
		if info.IsDir() {
			return os.MkdirAll(dest, os.ModePerm)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0644)
	})

	if copyErr != nil {
		// Rollback
		os.RemoveAll(projectCoreDir)
		if _, err := os.Stat(backupDir); err == nil {
			_ = os.Rename(backupDir, projectCoreDir)
		}
		return fmt.Errorf("failed to copy core files: %v", copyErr)
	}

	// Remove backup
	os.RemoveAll(backupDir)

	fmt.Println("Core directory updated successfully.")
	return nil
}
