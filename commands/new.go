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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var newCmd = &cobra.Command{
	Use:   "new [project_name]",
	Short: "Create a new project",
	Long:  `Create a new project by cloning the base repository and setting up the directory.`,
	Args:  cobra.ExactArgs(1),
	Run:   createNewProject,
}

func init() {
	rootCmd.AddCommand(newCmd)
}

func createNewProject(cmd *cobra.Command, args []string) {
	projectName := args[0]

	// Use the same version as the CLI for framework download
	// Normalize to ensure exactly one leading 'v' in the tag
	rawVersion := version.Version
	normalized := strings.TrimPrefix(rawVersion, "v")
	tag := "v" + normalized
	archiveURL := fmt.Sprintf("https://github.com/base-go/base-core/archive/refs/tags/%s.zip", tag)

	// Create the project directory
	err := os.Mkdir(projectName, 0755)
	if err != nil {
		fmt.Printf("Error creating project directory: %v\n", err)
		return
	}

	// Download the archive
	fmt.Printf("Downloading project template from: %s\n", archiveURL)
	resp, err := http.Get(archiveURL)
	if err != nil {
		fmt.Printf("Error downloading project template: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Check HTTP response status
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: HTTP %d when downloading template\n", resp.StatusCode)
		fmt.Printf("URL: %s\n", archiveURL)
		return
	}

	// Create a temporary file to store the zip
	tmpZip, err := os.CreateTemp("", "base-project-*.zip")
	if err != nil {
		fmt.Printf("Error creating temporary file: %v\n", err)
		return
	}
	defer os.Remove(tmpZip.Name())

	// Copy the zip content to the temporary file
	bytesWritten, err := io.Copy(tmpZip, resp.Body)
	if err != nil {
		fmt.Printf("Error saving project template: %v\n", err)
		return
	}
	tmpZip.Close()

	// Debug: Show file size and check if it's reasonable
	fmt.Printf("Downloaded %d bytes\n", bytesWritten)
	if bytesWritten < 1000 {
		fmt.Printf("Warning: Downloaded file seems too small (%d bytes)\n", bytesWritten)
		// Read the file to see what we actually downloaded
		content, _ := os.ReadFile(tmpZip.Name())
		fmt.Printf("Content preview: %s\n", string(content[:min(len(content), 200)]))
	}

	// Unzip the file
	err = utils.Unzip(tmpZip.Name(), projectName)
	if err != nil {
		fmt.Printf("Error extracting project template: %v\n", err)
		return
	}

	// Move contents from the version-specific subdirectory to the project root
	// GitHub archives may produce either base-vX.Y.Z or base-X.Y.Z
	candidateDirs := []string{
		fmt.Sprintf("base-%s", tag),        // e.g., base-v2.1.3
		fmt.Sprintf("base-%s", normalized), // e.g., base-2.1.3
	}

	var extractedDir string
	for _, d := range candidateDirs {
		p := filepath.Join(projectName, d)
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			extractedDir = p
			break
		}
	}

	// Fallback: pick the first directory that starts with "base-"
	if extractedDir == "" {
		if entries, err := os.ReadDir(projectName); err == nil {
			for _, e := range entries {
				if e.IsDir() && strings.HasPrefix(e.Name(), "base-") {
					extractedDir = filepath.Join(projectName, e.Name())
					break
				}
			}
		}
	}

	if extractedDir == "" {
		fmt.Printf("Error: could not locate extracted template directory. Looked for: %v\n", candidateDirs)
		if entries, err := os.ReadDir(projectName); err == nil {
			fmt.Printf("Contents of '%s':\n", projectName)
			for _, e := range entries {
				fmt.Println(" -", e.Name())
			}
		}
		return
	}

	files, err := os.ReadDir(extractedDir)
	if err != nil {
		fmt.Printf("Error reading template directory %s: %v\n", extractedDir, err)
		return
	}

	for _, f := range files {
		oldPath := filepath.Join(extractedDir, f.Name())
		newPath := filepath.Join(projectName, f.Name())
		err = os.Rename(oldPath, newPath)
		if err != nil {
			fmt.Printf("Error moving file %s: %v\n", f.Name(), err)
		}
	}

	// Remove the now-empty subdirectory
	os.RemoveAll(extractedDir)

	// Get the absolute path of the new project directory
	absPath, err := filepath.Abs(projectName)
	if err != nil {
		fmt.Printf("Error getting absolute path: %v\n", err)
		return
	}

	fmt.Printf("New project '%s' created successfully at %s\n", projectName, absPath)
	fmt.Println("You can now start working on your new project!")
}
