package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	outputDir string
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate Swagger documentation",
	Long:  `Generate Swagger documentation using swag by scanning controller annotations and create static files (JSON, YAML, docs.go).`,
	Run:   generateDocs,
}

func init() {
	docsCmd.Flags().StringVarP(&outputDir, "output", "o", "docs", "Output directory for generated files")
	rootCmd.AddCommand(docsCmd)
}

func generateDocs(cmd *cobra.Command, args []string) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting working directory: %v\n", err)
		return
	}

	// Check if we're in a Base project by looking for main.go
	mainPath := filepath.Join(cwd, "main.go")
	if _, err := os.Stat(mainPath); os.IsNotExist(err) {
		fmt.Println("Error: Base project structure not found.")
		fmt.Println("Make sure you are in the root directory of your Base project.")
		fmt.Println("Expected to find main.go at:", mainPath)
		return
	}

	// Find go executable using which
	whichCmd := exec.Command("which", "go")
	goPathBytes, err := whichCmd.Output()
	if err != nil {
		fmt.Printf("Error: Go executable not found: %v\n", err)
		fmt.Println("Please ensure Go is properly installed and in your PATH")
		return
	}
	goPath := strings.TrimSpace(string(goPathBytes))

	// Ensure swag is installed
	if _, err := exec.LookPath("swag"); err != nil {
		fmt.Println("Installing swag...")
		installCmd := exec.Command(goPath, "install", "github.com/swaggo/swag/cmd/swag@latest")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			fmt.Printf("Error installing swag: %v\n", err)
			return
		}
	}

	fmt.Println("📚 Generating swagger documentation from annotations...")

	swagCmd := exec.Command(
		"swag",
		"init",
		"--dir", "./",
		"--output", "./"+outputDir,
		"--parseDependency",
		"--parseInternal",
		"--parseVendor",
		"--parseDepth", "1",
		"--generatedTime", "false",
	)

	swagCmd.Dir = cwd
	swagCmd.Stdout = os.Stdout
	swagCmd.Stderr = os.Stderr

	if err := swagCmd.Run(); err != nil {
		fmt.Printf("Error generating docs: %v\n", err)
		return
	}

	fmt.Println("✅ Swagger documentation generated successfully!")
	fmt.Printf("   - %s/swagger.json\n", outputDir)
	fmt.Printf("   - %s/swagger.yaml\n", outputDir)
	fmt.Printf("   - %s/docs.go\n", outputDir)
}
