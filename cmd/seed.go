package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed the application",
	Long:  `Seed the application by running 'go run main.go seed' in the current directory.`,
	Run:   seedApplication,
}

var replantCmd = &cobra.Command{
	Use:   "replant",
	Short: "Clean and reseed the application",
	Long:  `Clean all tables and reseed the application by running 'go run main.go replant' in the current directory.`,
	Run:   replantApplication,
}

func init() {
	rootCmd.AddCommand(seedCmd)
	rootCmd.AddCommand(replantCmd)
}

func seedApplication(cmd *cobra.Command, args []string) {
	runMainWithArgument("seed")
}

func replantApplication(cmd *cobra.Command, args []string) {
	runMainWithArgument("replant")
}

func runMainWithArgument(argument string) {
	// Check if main.go exists in the current directory
	if _, err := os.Stat("main.go"); os.IsNotExist(err) {
		fmt.Println("Error: main.go not found in the current directory.")
		fmt.Println("Make sure you are in the root directory of your Base project.")
		return
	}

	// Run "go run main.go" with the given argument
	goCmd := exec.Command("go", "run", "main.go", argument)
	goCmd.Stdout = os.Stdout
	goCmd.Stderr = os.Stderr

	fmt.Printf("Running %s operation...\n", argument)
	err := goCmd.Run()
	if err != nil {
		fmt.Printf("Error running %s operation: %v\n", argument, err)
		return
	}
}
