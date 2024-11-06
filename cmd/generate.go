package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"base/utils"

	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:     "generate [name] [field:type...]",
	Aliases: []string{"g"},
	Short:   "Generate a new module",
	Long:    `Generate a new module with the specified name and fields.`,
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
	//pluralName := utils.ToPlural(singularName)
	//pluralDirName := utils.ToSnakeCase(singularName)

	// Use PascalCase for struct naming
	structName := utils.ToPascalCase(singularName)

	// Use the singular name in snake_case for package naming
	packageName := utils.ToSnakeCase(singularName)

	// Create directories (singular names in snake_case)
	dirs := []string{
		filepath.Join("app", "models"),
		filepath.Join("app", dirName), // Changed to snake_case singular directory
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			return
		}
	}

	// Process fields into FieldStruct
	processedFields := utils.GenerateFieldStructs(fields)

	// Generate model file with processed fields
	utils.GenerateFileFromTemplate(
		filepath.Join("app", "models"),
		dirName+".go",
		"templates/model.tmpl",
		structName,
		dirName,
		"models",
		processedFields,
	)

	// Generate other files (in singular directory with snake_case)
	files := []string{"controller.go", "service.go", "mod.go"}
	for _, file := range files {
		templateName := strings.TrimSuffix(file, ".go") + ".tmpl"
		utils.GenerateFileFromTemplate(
			filepath.Join("app", dirName), // Use singular directory in snake_case
			file,
			"templates/"+templateName,
			structName,
			dirName,
			packageName,
			processedFields,
		)
	}

	// Update app/init.go to register the new module
	if err := utils.UpdateInitFile(singularName); err != nil {
		fmt.Printf("Error updating app/init.go: %v\n", err)
		return
	}

	fmt.Printf("Module %s generated successfully with fields: %v\n", singularName, fields)
}
