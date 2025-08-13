package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/BaseTechStack/basecmd/utils"
	"github.com/BaseTechStack/basecmd/version"
	"github.com/spf13/cobra"
)

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
	cliVersion := version.Version
	archiveURL := fmt.Sprintf("https://github.com/BaseTechStack/base/archive/refs/tags/v%s.zip", cliVersion)

	// Create the project directory
	err := os.Mkdir(projectName, 0755)
	if err != nil {
		fmt.Printf("Error creating project directory: %v\n", err)
		return
	}

	// Download the archive
	resp, err := http.Get(archiveURL)
	if err != nil {
		fmt.Printf("Error downloading project template: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Create a temporary file to store the zip
	tmpZip, err := os.CreateTemp("", "base-project-*.zip")
	if err != nil {
		fmt.Printf("Error creating temporary file: %v\n", err)
		return
	}
	defer os.Remove(tmpZip.Name())

	// Copy the zip content to the temporary file
	_, err = io.Copy(tmpZip, resp.Body)
	if err != nil {
		fmt.Printf("Error saving project template: %v\n", err)
		return
	}
	tmpZip.Close()

	// Unzip the file
	err = utils.Unzip(tmpZip.Name(), projectName)
	if err != nil {
		fmt.Printf("Error extracting project template: %v\n", err)
		return
	}

	// Move contents from the version-specific subdirectory to the project root
	versionedDirName := fmt.Sprintf("base-%s", cliVersion)
	extractedDir := filepath.Join(projectName, versionedDirName)

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
