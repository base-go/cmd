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
	docs bool
)

var startCmd = &cobra.Command{
	Use:     "start",
	Aliases: []string{"s"},
	Short:   "Start the application",
	Long:    `Start the application by running the Base application server.`,
	Run:     startApplication,
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().BoolVarP(&docs, "docs", "d", false, "Generate Swagger documentation")
}

func startApplication(cmd *cobra.Command, args []string) {
	// Get the current working directory
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

	// Run go mod tidy to ensure dependencies are up to date
	fmt.Println("Ensuring dependencies are up to date...")
	tidyCmd := exec.Command(goPath, "mod", "tidy")
	tidyCmd.Dir = cwd
	if err := tidyCmd.Run(); err != nil {
		fmt.Printf("Warning: Failed to run go mod tidy: %v\n", err)
	}

	if docs {
		// Ensure swag is installed
		if _, err := exec.LookPath("swag"); err != nil {
			fmt.Println("Installing swag...")
			installCmd := exec.Command(goPath, "install", "github.com/swaggo/swag/cmd/swag@latest")
			installCmd.Stdout = os.Stdout
			installCmd.Stderr = os.Stderr
			if err := installCmd.Run(); err != nil {
				fmt.Printf("Warning: Failed to install swag: %v\n", err)
			}
		}

		// Generate swagger docs using swag
		swagCmd := exec.Command("swag", "init", "--dir", "./", "--output", "./docs", "--parseDependency", "--parseInternal", "--parseVendor", "--parseDepth", "1", "--generatedTime", "false")
		swagCmd.Dir = cwd
		swagCmd.Stdout = os.Stdout
		swagCmd.Stderr = os.Stderr

		if err := swagCmd.Run(); err != nil {
			fmt.Printf("Warning: Failed to generate docs: %v\n", err)
			fmt.Println("Continuing without auto-generated documentation...")
		}

		fmt.Println("📚 Swagger documentation will be available at /swagger/ when server starts")
	}

	// Run normally
	fmt.Println("Starting the Base application server...")

	mainCmd := exec.Command(goPath, "run", "main.go")
	mainCmd.Stdout = os.Stdout
	mainCmd.Stderr = os.Stderr
	mainCmd.Dir = cwd

	// Set environment variables
	env := os.Environ()
	if docs {
		env = append(env, "SWAGGER_ENABLED=true")
	}
	mainCmd.Env = env

	if err := mainCmd.Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		return
	}
}
