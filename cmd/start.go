package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
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
}

func startApplication(cmd *cobra.Command, args []string) {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
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

	// Run "go run main.go"
	fmt.Println("Starting the Base application server...")
	goCmd := exec.Command("go", "run", "main.go")
	goCmd.Stdout = os.Stdout
	goCmd.Stderr = os.Stderr
	goCmd.Dir = cwd

	if err := goCmd.Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		return
	}
}
