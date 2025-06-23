package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var packageCmd = &cobra.Command{
	Use:   "package",
	Short: "Manage Base packages",
	Long:  `Add or remove Base packages from your project.`,
}

var packageAddCmd = &cobra.Command{
	Use:   "add [namespace/package or URL]",
	Short: "Add a Base package",
	Long: `Add a Base package to your project.
You can specify either a namespace/package (e.g., base-packages/gamification)
or a full URL for non-GitHub packages.

Examples:
  # Add a package from base-packages
  base package add base-packages/gamification

  # Add a package using full URL
  base package add https://gitlab.com/org/package`,
	Run: addBasePackage,
}

var packageRemoveCmd = &cobra.Command{
	Use:   "remove [package-name]",
	Short: "Remove a Base package",
	Long: `Remove a Base package from your project.
This will remove the package files and its module initializer from start.go.

Example:
  base package remove gamification`,
	Run: removeBasePackage,
}

func init() {
	rootCmd.AddCommand(packageCmd)
	packageCmd.AddCommand(packageAddCmd)
	packageCmd.AddCommand(packageRemoveCmd)
}

func addBasePackage(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		listOfficialPackages()
		return
	}

	var repoURL string
	packageArg := args[0]

	// If it's a full URL, use it directly
	if strings.HasPrefix(packageArg, "http") {
		repoURL = packageArg
	} else {
		// If it's namespace/package format, construct GitHub URL
		parts := strings.Split(packageArg, "/")
		if len(parts) != 2 {
			fmt.Println("Error: Invalid package format. Use 'namespace/package' or a full URL")
			return
		}
		repoURL = fmt.Sprintf("https://github.com/%s", packageArg)
	}

	fmt.Printf("Adding package from %s...\n", repoURL)
	err := addPackage(repoURL)
	if err != nil {
		fmt.Printf("Error adding package: %v\n", err)
		return
	}
	fmt.Println("Package added successfully.")
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
	markerStr := "// PACKAGE INITIALIZER MARKER - Do not remove this comment because it's used by the CLI to add new package initializers"
	markerIndex := strings.Index(string(content), markerStr)
	if markerIndex == -1 {
		return fmt.Errorf("package initializer marker not found in start.go")
	}

	// Convert package name to PascalCase for the struct name
	structName := cases.Title(language.AmericanEnglish).String(packageName)

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

	// Write changes and run goimports
	if err := os.WriteFile(startFile, content, 0644); err != nil {
		return fmt.Errorf("failed to write start.go: %v", err)
	}

	cmd := exec.Command("goimports", "-w", startFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to run goimports: %v\n%s", err, output)
	}

	return nil
}

func removeBasePackage(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Package name is required")
		return
	}

	packageName := args[0]
	fmt.Printf("Removing package %s...\n", packageName)

	// Remove package directory
	packageDir := filepath.Join(".", "packages", packageName)
	if err := os.RemoveAll(packageDir); err != nil {
		fmt.Printf("Error removing package directory: %v\n", err)
		return
	}

	// Remove module initializer from start.go
	if err := removeModuleInitializer(packageName); err != nil {
		fmt.Printf("Error removing module initializer: %v\n", err)
		return
	}

	fmt.Println("Package removed successfully.")
}

func removeModuleInitializer(packageName string) error {
	startFile := filepath.Join(".", "core", "start.go")
	content, err := os.ReadFile(startFile)
	if err != nil {
		return fmt.Errorf("failed to read start.go: %v", err)
	}

	// Remove import
	importStr := fmt.Sprintf(`"base/packages/%s"`, packageName)
	contentStr := string(content)
	if idx := strings.Index(contentStr, importStr); idx != -1 {
		// Find the line start and end
		lineStart := strings.LastIndex(contentStr[:idx], "\n") + 1
		lineEnd := strings.Index(contentStr[idx:], "\n") + idx
		if lineEnd == -1 {
			lineEnd = len(contentStr)
		}
		contentStr = contentStr[:lineStart] + contentStr[lineEnd+1:]
	}

	// Remove initializer block
	structName := strings.Title(packageName)
	initializerStart := fmt.Sprintf("// Initialize %s modules", structName)
	if idx := strings.Index(contentStr, initializerStart); idx != -1 {
		// Find the block end (next empty line)
		blockEnd := strings.Index(contentStr[idx:], "\n\n")
		if blockEnd == -1 {
			blockEnd = len(contentStr)
		} else {
			blockEnd += idx + 1 // Include one newline
		}
		contentStr = contentStr[:idx] + contentStr[blockEnd:]
	}

	// Write changes and run goimports
	if err := os.WriteFile(startFile, []byte(contentStr), 0644); err != nil {
		return fmt.Errorf("failed to write start.go: %v", err)
	}

	cmd := exec.Command("goimports", "-w", startFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to run goimports: %v\n%s", err, output)
	}

	return nil
}

type Package struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

func listOfficialPackages() {
	packages := []Package{
		{
			Name:        "gamification",
			Description: "Gamification package for Base applications",
			URL:         "https://github.com/base-packages/gamification",
		},
		{
			Name:        "blog",
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
		fmt.Printf("  URL: %s\n", pkg.URL)
		fmt.Printf("  Install with: base package add base-packages/%s\n\n", pkg.Name)
	}
}
