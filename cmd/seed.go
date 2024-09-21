package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var seedCmd = &cobra.Command{
	Use:     "seed [seedName]",
	Aliases: []string{"s"},
	Short:   "Run database seeds",
	Long:    `Run database seeds. Use 'all' to run all seeders or specify a seeder name.`,
	Run:     runSeed,
}

func init() {
	rootCmd.AddCommand(seedCmd)
}

func runSeed(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Seed name is required. Use 'all' to run all seeders or specify a seeder name.")
		return
	}

	seedName := args[0]

	// Check if the seeds directory exists
	seedsDir := "app/seeds"
	if _, err := os.Stat(seedsDir); os.IsNotExist(err) {
		fmt.Printf("Error: %s directory not found.\n", seedsDir)
		fmt.Println("Make sure you are in the root directory of your project.")
		return
	}

	// Run the internal seed command
	seedCmd := exec.Command(os.Args[0], "internal-seed", seedName)
	seedCmd.Stdout = os.Stdout
	seedCmd.Stderr = os.Stderr
	seedCmd.Dir = filepath.Dir(seedsDir) // Set working directory to the project root

	fmt.Println("Running seeds...")
	err := seedCmd.Run()
	if err != nil {
		fmt.Printf("Error running seeds: %v\n", err)
		return
	}

	fmt.Println("Seeds executed successfully.")
}
