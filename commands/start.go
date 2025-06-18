package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/BaseTechStack/basecmd/utils"
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

func ensureSwagInstalled() error {
	if _, err := exec.LookPath("swag"); err != nil {
		fmt.Println("Installing swag for API documentation...")
		cmd := exec.Command("go", "install", "github.com/swaggo/swag/cmd/swag@latest")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	return nil
}

func generateSwaggerDocs(cwd string) error {
	fmt.Println("Generating Swagger documentation...")
	cmd := exec.Command("swag", "init",
		"--parseDependency",
		"--parseInternal",
		"--parseVendor",
		"--parseDepth", "1",
		"--generatedTime=false")
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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

	if docs {
		// Install swag if needed
		if err := ensureSwagInstalled(); err != nil {
			fmt.Printf("Error installing swag: %v\n", err)
			return
		}

		// Generate swagger docs
		if err := generateSwaggerDocs(cwd); err != nil {
			fmt.Printf("Error generating swagger docs: %v\n", err)
			return
		}
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
		airCmd.Env = append(os.Environ(), "SWAG_DISABLED=true")

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

		if err := mainCmd.Run(); err != nil {
			fmt.Printf("Error running application: %v\n", err)
			return
		}
	}
}
