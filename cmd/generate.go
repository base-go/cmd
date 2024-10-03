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
	Use:     "generate [name] [field:type...] [--admin]",
	Aliases: []string{"g"},
	Short:   "Generate a new module",
	Long:    `Generate a new module with the specified name and fields. Use --admin flag to generate admin interface.`,
	Args:    cobra.MinimumNArgs(1),
	Run:     generateModule,
}

func init() {
	generateCmd.Flags().Bool("admin", false, "Generate admin interface")
	rootCmd.AddCommand(generateCmd)
}

// generateModule generates a new module with the specified name and fields.
func generateModule(cmd *cobra.Command, args []string) {
	singularName := args[0]
	fields := args[1:]

	// Convert singular name to snake_case for directory naming
	dirName := utils.ToSnakeCase(singularName)
	pluralName := utils.ToPlural(singularName)
	pluralDirName := utils.ToSnakeCase(pluralName)

	// Use PascalCase for struct naming
	structName := utils.ToPascalCase(singularName)

	// Use the singular name in snake_case for package naming
	packageName := utils.ToSnakeCase(singularName)

	// Create directories (singular names in snake_case)
	dirs := []string{
		filepath.Join("app", "models"),
		filepath.Join("app", dirName),
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
	files := []string{"controller.go", "service.go", "mod.go", "seed.go"}
	for _, file := range files {
		templateName := strings.TrimSuffix(file, ".go") + ".tmpl"
		utils.GenerateFileFromTemplate(
			filepath.Join("app", dirName),
			file,
			"templates/"+templateName,
			structName,
			pluralDirName,
			packageName,
			processedFields,
		)
	}

	// Update app/init.go to register the new module
	if err := utils.UpdateInitFile(singularName, pluralName); err != nil {
		fmt.Printf("Error updating app/init.go: %v\n", err)
		return
	}

	// Update seeders in app/seed.go
	if err := utils.UpdateSeedersFile(structName, packageName); err != nil {
		fmt.Printf("Error updating seeders in app/seed.go: %v\n", err)
		return
	}

	adminFlag, _ := cmd.Flags().GetBool("admin")
	if adminFlag {
		generateAdminInterface(singularName, pluralName, fields)
	}

	fmt.Printf("Module %s generated successfully with fields: %v\n", singularName, fields)
}

func generateAdminInterface(singularName, pluralName string, fields []string) {
	pluralSnakeCase := utils.ToSnakeCase(pluralName)
	adminDir := filepath.Join("admin", pluralSnakeCase)
	if err := os.MkdirAll(adminDir, os.ModePerm); err != nil {
		fmt.Printf("Error creating admin directory %s: %v\n", adminDir, err)
		return
	}

	adminTemplateContent, err := utils.TemplateFS.ReadFile("templates/admin_interface.tmpl")
	if err != nil {
		fmt.Printf("Error reading admin template: %v\n", err)
		return
	}

	fieldStructs := utils.GenerateFieldStructs(fields)

	data := map[string]interface{}{
		"StructName": utils.ToTitle(singularName),
		"PluralName": utils.ToTitle(pluralName),
		"RouteName":  pluralSnakeCase,
		"Fields":     fieldStructs,
	}

	tmpl, err := utils.ParseTemplate("admin_interface", string(adminTemplateContent))
	if err != nil {
		fmt.Printf("Error parsing admin template: %v\n", err)
		return
	}

	filePath := filepath.Join(adminDir, "index.html")
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", filePath, err)
		return
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		fmt.Printf("Error executing template for index.html: %v\n", err)
		return
	}

	// Update admin/partials/nav.html
	utils.UpdateNavFile(pluralSnakeCase)

	// Update admin/index.html
	utils.UpdateIndexFile(pluralSnakeCase)

	fmt.Printf("Admin interface for %s generated in %s\n", singularName, adminDir)
}
