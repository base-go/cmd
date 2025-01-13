package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"base/utils"

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

	// Check if the module exists
	moduleDir := filepath.Join("app", pluralName)
	if _, err := os.Stat(moduleDir); os.IsNotExist(err) {
		fmt.Printf("Module '%s' does not exist.\n", singularName)
		return
	}

	// Delete module directory
	if err := os.RemoveAll(moduleDir); err != nil {
		fmt.Printf("Error removing directory %s: %v\n", moduleDir, err)
		return
	}

	// Delete model file
	modelFile := filepath.Join("app", "models", utils.ToSnakeCase(singularName)+".go")
	if err := os.Remove(modelFile); err != nil {
		fmt.Printf("Error removing model file %s: %v\n", modelFile, err)
		return
	}

	// Update app/init.go to unregister the module
	if err := utils.UpdateInitFileForDestroy(pluralName); err != nil {
		fmt.Printf("Error updating app/init.go: %v\n", err)
		return
	}

	fmt.Printf("Successfully destroyed module '%s':\n", singularName)
	fmt.Printf("- Removed directory: %s\n", moduleDir)
	fmt.Printf("- Removed model file: app/models/%s.go\n", utils.ToSnakeCase(singularName))
	fmt.Printf("- Updated app/init.go\n")
}
