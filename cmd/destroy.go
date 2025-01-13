package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/base-go/cmd/utils"
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

	fmt.Printf("Successfully destroyed module '%s'.\n", singularName)
}
