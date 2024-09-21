package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var seedCmd = &cobra.Command{
	Use:     "seed",
	Aliases: []string{"s"},
	Short:   "Run database seeds",
	Long:    `Run all database seeds.`,
	Run:     runSeed,
}

var projectRoot string

func init() {
	rootCmd.AddCommand(seedCmd)
	seedCmd.Flags().StringVarP(&projectRoot, "project-root", "p", "", "Path to the project root")
}

func runSeed(cmd *cobra.Command, args []string) {
	if projectRoot == "" {
		var err error
		projectRoot, err = os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			return
		}
	}

	seedsPath := filepath.Join(projectRoot, "app", "seeds")
	if _, err := os.Stat(seedsPath); os.IsNotExist(err) {
		fmt.Printf("Error: Seeds directory not found at %s\n", seedsPath)
		return
	}

	// Run the go run command
	goCmd := exec.Command("go", "run", filepath.Join(seedsPath, "all.go"))
	goCmd.Stdout = cmd.OutOrStdout()
	goCmd.Stderr = cmd.ErrOrStderr()
	goCmd.Dir = projectRoot

	err := goCmd.Run()
	if err != nil {
		fmt.Printf("Error running seeds: %v\n", err)
		return
	}

	fmt.Println("Seeds executed successfully.")
}
