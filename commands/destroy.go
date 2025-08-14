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
	var moduleExists bool

	pluralDir := filepath.Join("app", pluralName)
	singularDir := filepath.Join("app", singularDirName)

	if _, err := os.Stat(pluralDir); err == nil {
		moduleDir = pluralDir
		moduleExists = true
	} else if _, err := os.Stat(singularDir); err == nil {
		moduleDir = singularDir
		moduleExists = true
	} else {
		// Module directory doesn't exist, but we can still clean up orphaned entries
		fmt.Printf("Module directory '%s' does not exist, but checking for orphaned entries in init.go...\n", singularName)
		moduleExists = false
	}

	// Prompt for confirmation with Y preselected
	reader := bufio.NewReader(os.Stdin)
	if moduleExists {
		fmt.Printf("Are you sure you want to destroy the '%s' module? [Y/n] ", singularName)
	} else {
		fmt.Printf("Are you sure you want to clean up orphaned '%s' entries from init.go? [Y/n] ", singularName)
	}
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

	// Delete module directory if it exists
	if moduleExists {
		if err := os.RemoveAll(moduleDir); err != nil {
			fmt.Printf("Error removing directory %s: %v\n", moduleDir, err)
			return
		}
		fmt.Printf("Removed module directory: %s\n", moduleDir)
	}

	// Delete model file if module exists - check both models and model directories
	if moduleExists {
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
	}

	// Delete test files if module exists
	if moduleExists {
		testDir := filepath.Join("test", "app_test", pluralName+"_test")
		if _, err := os.Stat(testDir); err == nil {
			if err := os.RemoveAll(testDir); err != nil {
				fmt.Printf("Warning: Could not remove test directory %s: %v\n", testDir, err)
			} else {
				fmt.Printf("Removed test directory: %s\n", testDir)
			}
		}
	}

	// Remove import from base/app/init.go
	if err := removeModuleFromAppInit(pluralName); err != nil {
		fmt.Printf("Warning: Could not remove import from base/app/init.go: %v\n", err)
	} else {
		fmt.Printf("✅ Removed '%s' from base/app/init.go\n", pluralName)
	}

	fmt.Printf("Successfully destroyed module '%s'.\n", singularName)
}

// removeModuleFromAppInit removes the module from app/init.go
func removeModuleFromAppInit(moduleName string) error {
	initGoPath := filepath.Join("app", "init.go")

	// Check if app/init.go exists
	if _, err := os.Stat(initGoPath); os.IsNotExist(err) {
		return nil // Nothing to remove
	}

	// Read the file
	content, err := os.ReadFile(initGoPath)
	if err != nil {
		return fmt.Errorf("failed to read app/init.go: %w", err)
	}

	// Remove the import line using the helper function
	importStr := fmt.Sprintf("\"base/app/%s\"", moduleName)
	content = utils.RemoveImport(content, importStr)

	// Remove the module initialization using the helper function
	content = utils.RemoveModuleInitializer(content, moduleName)

	// Write back to file
	if err := os.WriteFile(initGoPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write app/init.go: %w", err)
	}

	return nil
}
