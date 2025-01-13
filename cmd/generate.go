package cmd

import (
	"base/utils"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:     "generate [name] [field:type...]",
	Aliases: []string{"g"},
	Short:   "Generate a new module",
	Long:    `Generate a new module with the specified name and fields. Use --admin flag to generate admin interface.`,
	Args:    cobra.MinimumNArgs(1),
	Run:     generateModule,
}

func init() {
	rootCmd.AddCommand(generateCmd)
}

// generateModule generates a new module with the specified name and fields.
func generateModule(cmd *cobra.Command, args []string) {
	singularName := args[0]
	fields := args[1:]

	// Convert singular name to snake_case for directory naming
	dirName := utils.ToSnakeCase(singularName)
	pluralName := utils.PluralizeClient.Plural(singularName)
	pluralDirName := utils.ToSnakeCase(pluralName)

	// Use PascalCase for struct naming
	structName := utils.ToPascalCase(singularName)

	// Use the plural name in snake_case for package naming
	packageName := utils.ToSnakeCase(pluralName)

	// Create directories (plural names in snake_case)
	dirs := []string{
		filepath.Join("app", "models"),
		filepath.Join("app", pluralDirName),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			return
		}
	}

	// Generate field structs
	fieldStructs := utils.GenerateFieldStructs(fields)

	// Generate model
	utils.GenerateFileFromTemplate(
		filepath.Join("app", "models"),
		fmt.Sprintf("%s.go", dirName),
		"model.tmpl",
		structName,
		pluralName,
		"models",
		fieldStructs,
	)

	// Generate service
	utils.GenerateFileFromTemplate(
		filepath.Join("app", pluralDirName),
		"service.go",
		"service.tmpl",
		structName,
		pluralName,
		packageName,
		fieldStructs,
	)

	// Generate controller
	utils.GenerateFileFromTemplate(
		filepath.Join("app", pluralDirName),
		"controller.go",
		"controller.tmpl",
		structName,
		pluralName,
		packageName,
		fieldStructs,
	)

	// Generate module
	utils.GenerateFileFromTemplate(
		filepath.Join("app", pluralDirName),
		"module.go",
		"module.tmpl",
		structName,
		pluralName,
		packageName,
		fieldStructs,
	)

	// Update init.go
	if err := utils.UpdateInitGo(pluralDirName, structName); err != nil {
		fmt.Printf("Error updating init.go: %v\n", err)
		return
	}
}
