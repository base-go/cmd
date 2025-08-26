package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/base-go/cmd/utils"
	"github.com/spf13/cobra"
)

var (
	hotReload bool
	docs      bool
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
	startCmd.Flags().BoolVarP(&hotReload, "hot-reload", "r", false, "Enable hot reloading using air")
	startCmd.Flags().BoolVarP(&docs, "docs", "d", false, "Generate Swagger documentation")
}

func ensureAirInstalled() error {
	// Check if air is installed
	if _, err := exec.LookPath("air"); err != nil {
		fmt.Println("Installing air for hot reloading...")
		cmd := exec.Command("go", "install", "github.com/air-verse/air@latest")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	return nil
}

func setupAirConfig(cwd string) error {
	airConfigPath := filepath.Join(cwd, ".air.toml")

	// Only create config if it doesn't exist
	if _, err := os.Stat(airConfigPath); os.IsNotExist(err) {
		fmt.Println("Creating air configuration...")
		if err := utils.GenerateAirFileFromTemplate(cwd); err != nil {
			return fmt.Errorf("failed to generate air config: %w", err)
		}
	}
	return nil
}

// Base framework uses custom swagger implementation
// No need for swaggo/swag installation or generation

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

	// Run go mod tidy to ensure dependencies are up to date
	fmt.Println("Ensuring dependencies are up to date...")
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = cwd
	if err := tidyCmd.Run(); err != nil {
		fmt.Printf("Warning: Failed to run go mod tidy: %v\n", err)
	}

	if docs {
		fmt.Println("📚 Generating swagger documentation from annotations...")

		// Generate swagger docs using the new docs command
		docsCmd := exec.Command("base", "docs")
		docsCmd.Dir = cwd
		docsCmd.Stdout = os.Stdout
		docsCmd.Stderr = os.Stderr

		if err := docsCmd.Run(); err != nil {
			fmt.Printf("Warning: Failed to generate docs: %v\n", err)
			fmt.Println("Continuing without auto-generated documentation...")
		}

		fmt.Println("📚 Swagger documentation will be available at /swagger/ when server starts")
	}

	if hotReload {
		// Install air if needed
		if err := ensureAirInstalled(); err != nil {
			fmt.Printf("Error installing air: %v\n", err)
			return
		}

		// Setup air config
		if err := setupAirConfig(cwd); err != nil {
			fmt.Printf("Error setting up air config: %v\n", err)
			return
		}

		// Run with air
		fmt.Println("Starting the Base application server with hot reloading...")
		airCmd := exec.Command("air")
		airCmd.Stdout = os.Stdout
		airCmd.Stderr = os.Stderr
		airCmd.Dir = cwd

		// Set environment variables
		env := os.Environ()
		if docs {
			env = append(env, "SWAGGER_ENABLED=true")
		}
		airCmd.Env = env

		if err := airCmd.Run(); err != nil {
			fmt.Printf("Error running application with air: %v\n", err)
			return
		}
	} else {
		// Run normally
		fmt.Println("Starting the Base application server...")
		fmt.Println("Tip: Use --hot-reload or -r flag to enable hot reloading")

		mainCmd := exec.Command("go", "run", "main.go")
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
}
