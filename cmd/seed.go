package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "seed the application",
	Long:  `seed the application by running 'go run main.go' in the current directory.`,
	Run:   seedApplication,
}

func init() {
	rootCmd.AddCommand(seedCmd)
}

func seedApplication(cmd *cobra.Command, args []string) {
	// Check if main.go exists in the current directory
	if _, err := os.Stat("main.go"); os.IsNotExist(err) {
		fmt.Println("Error: main.go not found in the current directory.")
		fmt.Println("Make sure you are in the root directory of your Base project.")
		return
	}

	// Run "go run main.go"
	goCmd := exec.Command("go", "run", "main.go seed")
	goCmd.Stdout = os.Stdout
	goCmd.Stderr = os.Stderr

	fmt.Println("seeding the application...")
	err := goCmd.Run()
	if err != nil {
		fmt.Printf("Error seeding the application: %v\n", err)
		return
	}
}
