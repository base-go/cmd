package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:     "start",
	Aliases: []string{"s"},
	Short:   "Start the application",
	Long:    `Start the application by running 'go run main.go' in the current directory.`,
	Run:     startApplication,
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func startApplication(cmd *cobra.Command, args []string) {
	// Check if main.go exists in the current directory
	if _, err := os.Stat("main.go"); os.IsNotExist(err) {
		fmt.Println("Error: main.go not found in the current directory.")
		fmt.Println("Make sure you are in the root directory of your Base project.")
		return
	}

	// Run "go run main.go"
	goCmd := exec.Command("go", "run", "main.go")
	goCmd.Stdout = os.Stdout
	goCmd.Stderr = os.Stderr

	fmt.Println("Starting the application...")
	err := goCmd.Run()
	if err != nil {
		fmt.Printf("Error starting the application: %v\n", err)
		return
	}
}
