package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [package URL]",
	Short: "add Base premade package",
	Long: `add Base premade packages from the community.
If no URL is provided, it will list the official packages.
You can add packages directly from GitHub using URLs like:
  base add https://github.com/Base-Packages/gamification`,
	Run: addBasePackages,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

type Package struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

func addBasePackages(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		listOfficialPackages()
		return
	}

	packageURL := args[0]
	if !strings.HasPrefix(packageURL, "https://github.com/") {
		fmt.Println("Error: Only GitHub URLs are supported")
		return
	}

	fmt.Printf("Adding package from %s...\n", packageURL)
	err := addPackage(packageURL)
	if err != nil {
		fmt.Printf("Error adding package: %v\n", err)
		return
	}
	fmt.Println("Package added successfully.")
}

func listOfficialPackages() {
	packages := []Package{
		{
			Name:        "gamification",
			Description: "Gamification package for Base applications",
			URL:         "https://github.com/base-packages/gamification",
		},
		{
			Name:        "Blog",
			Description: "Blog package for Base applications",
			URL:         "https://github.com/base-packages/blog",
		},
		// Add more official packages here
	}

	fmt.Println("Available official packages:")
	fmt.Println("---------------------------")
	for _, pkg := range packages {
		fmt.Printf("• %s\n", pkg.Name)
		fmt.Printf("  Description: %s\n", pkg.Description)
		fmt.Printf("  URL: %s\n\n", pkg.URL)
	}
	fmt.Println("To add a package, run: base add <package-url>")
}

func addPackage(repoURL string) error {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "base-package-add")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Clone the repository
	gitCmd := exec.Command("git", "clone", repoURL, tempDir)
	if output, err := gitCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to clone repository: %v\n%s", err, output)
	}

	// Get the package name from the URL
	urlParts := strings.Split(repoURL, "/")
	packageName := urlParts[len(urlParts)-1]

	// Path to the package directory in the current project
	projectPackageDir := filepath.Join(".", "packages", packageName)

	// Create packages directory if it doesn't exist
	if err := os.MkdirAll(filepath.Join(".", "packages"), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create packages directory: %v", err)
	}

	// Remove existing package directory if it exists
	os.RemoveAll(projectPackageDir)

	// Copy package files from temp directory to the project
	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		relPath, err := filepath.Rel(tempDir, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(projectPackageDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, os.ModePerm)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(targetPath, data, 0644)
	})

	if err != nil {
		return fmt.Errorf("failed to copy package files: %v", err)
	}

	// Add module initializer to start.go
	if err := addModuleInitializer(packageName); err != nil {
		return fmt.Errorf("failed to add module initializer: %v", err)
	}

	return nil
}

func addModuleInitializer(packageName string) error {
	startFile := filepath.Join(".", "core", "start.go")
	content, err := os.ReadFile(startFile)
	if err != nil {
		return fmt.Errorf("failed to read start.go: %v", err)
	}

	// Add import if not exists
	importStr := fmt.Sprintf(`	"base/core/packages/%s"`, packageName)
	if !strings.Contains(string(content), importStr) {
		importIndex := strings.Index(string(content), ")")
		if importIndex == -1 {
			return fmt.Errorf("invalid start.go file format")
		}
		content = []byte(string(content[:importIndex]) + importStr + "\n" + string(content[importIndex:]))
	}

	// Add initializer if not exists
	markerStr := "// PACKAGE INITIALIZER MARKER"
	markerIndex := strings.Index(string(content), markerStr)
	if markerIndex == -1 {
		return fmt.Errorf("package initializer marker not found in start.go")
	}

	// Convert package name to PascalCase for the struct name
	structName := strings.Title(packageName)
	
	initializerCode := fmt.Sprintf(`
	// Initialize %s modules
	%sInitializer := &%s.%sModuleInitializer{
		Router:  protectedGroup,
		Logger:  appLogger,
		Emitter: Emitter,
		Storage: activeStorage,
	}
	%sInitializer.InitializeModules(db.DB)

	%s`, 
		structName, packageName, packageName, structName, packageName, markerStr)

	// Replace the marker with the new initializer code
	content = []byte(strings.Replace(string(content), markerStr, initializerCode, 1))

	// Write the changes
	if err := os.WriteFile(startFile, content, 0644); err != nil {
		return fmt.Errorf("failed to write start.go: %v", err)
	}

	// Run goimports to fix imports
	cmd := exec.Command("goimports", "-w", startFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to run goimports: %v\n%s", err, output)
	}

	return nil
}
