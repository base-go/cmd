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
	dirName := utils.ToSnakeCase(singularName)

	// Check if the module exists
	moduleDir := filepath.Join("app", dirName)
	if _, err := os.Stat(moduleDir); os.IsNotExist(err) {
		fmt.Printf("Module '%s' does not exist.\n", singularName)
		return
	}

	// Delete module directory
	err := os.RemoveAll(moduleDir)
	if err != nil {
		fmt.Printf("Error removing directory %s: %v\n", moduleDir, err)
		return
	}
	// Remove model file
	err = os.Remove(filepath.Join("app", "models", singularName+".go"))
	if err != nil {
		fmt.Printf("Error removing model file: %v\n", err)
		return
	}

	// Update app/init.go to unregister the module
	err = utils.UpdateInitFileForDestroy(dirName)
	if err != nil {
		fmt.Printf("Error updating app/init.go: %v\n", err)
		return
	}

	fmt.Printf("Module '%s' destroyed successfully.\n", singularName)
	fmt.Println("Module unregistered from app/init.go successfully!")
}
