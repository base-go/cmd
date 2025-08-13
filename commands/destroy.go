package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BaseTechStack/basecmd/utils"
	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:   "d [name]",
	Short: "Destroy an existing module",
	Long:  `Destroy an existing module with the specified name.`,
	Args:  cobra.ExactArgs(1),
	Run:   destroyModule,
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}

func destroyModule(cmd *cobra.Command, args []string) {
	singularName := args[0]
	pluralName := utils.ToSnakeCase(utils.ToPlural(singularName))
	singularDirName := utils.ToSnakeCase(singularName)

	// Check if the module exists - try both plural and singular directory names
	var moduleDir string
	
	pluralDir := filepath.Join("app", pluralName)
	singularDir := filepath.Join("app", singularDirName)
	
	if _, err := os.Stat(pluralDir); err == nil {
		moduleDir = pluralDir
	} else if _, err := os.Stat(singularDir); err == nil {
		moduleDir = singularDir
	} else {
		fmt.Printf("Module '%s' does not exist.\n", singularName)
		return
	}

	// Prompt for confirmation with Y preselected
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Are you sure you want to destroy the '%s' module? [Y/n] ", singularName)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		return
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "" && response != "y" {
		fmt.Println("Operation cancelled.")
		return
	}

	// Delete module directory
	if err := os.RemoveAll(moduleDir); err != nil {
		fmt.Printf("Error removing directory %s: %v\n", moduleDir, err)
		return
	}

	// Delete model file - check both models and model directories
	modelFiles := []string{
		filepath.Join("app", "models", utils.ToSnakeCase(singularName)+".go"),
		filepath.Join("app", "model", utils.ToSnakeCase(singularName)+".go"),
	}
	
	modelRemoved := false
	for _, modelFile := range modelFiles {
		if err := os.Remove(modelFile); err == nil {
			fmt.Printf("Removed model file: %s\n", modelFile)
			modelRemoved = true
			break
		}
	}
	
	if !modelRemoved {
		fmt.Printf("Warning: No model file found for %s\n", singularName)
	}

	// Delete test files
	testDir := filepath.Join("test", "app_test", pluralName+"_test")
	if _, err := os.Stat(testDir); err == nil {
		if err := os.RemoveAll(testDir); err != nil {
			fmt.Printf("Warning: Could not remove test directory %s: %v\n", testDir, err)
		} else {
			fmt.Printf("Removed test directory: %s\n", testDir)
		}
	}

	fmt.Printf("Successfully destroyed module '%s'.\n", singularName)
}
