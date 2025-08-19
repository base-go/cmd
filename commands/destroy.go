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
	Use:     "d [name1] [name2] ...",
	Aliases: []string{"destroy"},
	Short:   "Destroy existing modules",
	Long:    `Destroy one or more existing modules with the specified names.`,
	Args:    cobra.MinimumNArgs(1),
	Run:     destroyModule,
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}

func destroyModule(cmd *cobra.Command, args []string) {
	// Show summary of modules to be destroyed
	fmt.Printf("Modules to destroy: %s\n", strings.Join(args, ", "))
	
	// Confirm destruction of all modules
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Are you sure you want to destroy %d module(s)? [Y/n] ", len(args))
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

	// Process each module
	allSuccessful := true
	for i, moduleName := range args {
		fmt.Printf("\n[%d/%d] Destroying module '%s'...\n", i+1, len(args), moduleName)
		if !destroySingleModule(moduleName) {
			allSuccessful = false
		}
	}

	if allSuccessful {
		fmt.Printf("\n✅ Successfully destroyed all %d module(s).\n", len(args))
	} else {
		fmt.Printf("\n⚠️  Some modules could not be fully destroyed. Check the output above for details.\n")
	}
}

func destroySingleModule(singularName string) bool {
	pluralName := utils.ToSnakeCase(utils.ToPlural(singularName))
	singularDirName := utils.ToSnakeCase(singularName)

	// Check if the module exists - try both plural and singular directory names
	var moduleDir string
	var moduleExists bool
	success := true

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
		fmt.Printf("  Module directory '%s' does not exist, checking for orphaned entries...\n", singularName)
		moduleExists = false
	}

	// Delete module directory if it exists
	if moduleExists {
		if err := os.RemoveAll(moduleDir); err != nil {
			fmt.Printf("  ❌ Error removing directory %s: %v\n", moduleDir, err)
			success = false
		} else {
			fmt.Printf("  ✅ Removed module directory: %s\n", moduleDir)
		}
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
				fmt.Printf("  ✅ Removed model file: %s\n", modelFile)
				modelRemoved = true
				break
			}
		}

		if !modelRemoved {
			fmt.Printf("  ⚠️  Warning: No model file found for %s\n", singularName)
		}
	}

	// Delete test files if module exists
	if moduleExists {
		testDir := filepath.Join("test", "app_test", pluralName+"_test")
		if _, err := os.Stat(testDir); err == nil {
			if err := os.RemoveAll(testDir); err != nil {
				fmt.Printf("  ⚠️  Warning: Could not remove test directory %s: %v\n", testDir, err)
			} else {
				fmt.Printf("  ✅ Removed test directory: %s\n", testDir)
			}
		}
	}

	// Remove import from base/app/init.go
	if err := removeModuleFromAppInit(pluralName); err != nil {
		fmt.Printf("  ⚠️  Warning: Could not remove import from base/app/init.go: %v\n", err)
	} else {
		fmt.Printf("  ✅ Removed '%s' from base/app/init.go\n", pluralName)
	}

	if success {
		fmt.Printf("  ✅ Successfully destroyed module '%s'\n", singularName)
	}
	
	return success
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
