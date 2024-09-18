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
	Use:   "g [name] [field:type...] [--admin]",
	Short: "Generate a new module",
	Long:  `Generate a new module with the specified name and fields. Use --admin flag to generate admin interface.`,
	Args:  cobra.MinimumNArgs(1),
	Run:   generateModule,
}

func init() {
	generateCmd.Flags().Bool("admin", false, "Generate admin interface")
	rootCmd.AddCommand(generateCmd)
}

func generateModule(cmd *cobra.Command, args []string) {
	singularName := args[0]
	fields := args[1:]

	// Convert singular name to snake_case for directory naming
	dirName := utils.ToSnakeCase(singularName)
	pluralDirName := utils.ToSnakeCase(utils.ToPlural(singularName))

	// Use PascalCase for struct naming
	structName := utils.ToPascalCase(singularName)

	// Use the singular name in snake_case for package naming
	packageName := utils.ToSnakeCase(singularName)

	// Create directories
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

	// Generate model file
	utils.GenerateFileFromTemplate(
		filepath.Join("app", "models"),
		dirName+".go",
		"templates/model.tmpl",
		structName,
		pluralDirName,
		"models",
		fields,
	)

	// Generate other files
	files := []string{"controller.go", "service.go", "mod.go"}
	for _, file := range files {
		utils.GenerateFileFromTemplate(
			filepath.Join("app", singularName),
			file,
			"templates/"+strings.TrimSuffix(file, ".go")+".tmpl",
			structName,
			pluralDirName,
			packageName,
			fields,
		)
	}

	// Update app/init.go to register the new module
	if err := utils.UpdateInitFile(singularName, utils.ToLowerPlural(structName)); err != nil {
		fmt.Printf("Error updating app/init.go: %v\n", err)
		return
	}

	adminFlag, _ := cmd.Flags().GetBool("admin")
	if adminFlag {
		generateAdminInterface(singularName, structName, fields)
	}

	fmt.Printf("Module %s generated successfully with fields: %v\n", singularName, fields)
}

func generateAdminInterface(singularName, pluralName string, fields []string) {
	adminDir := filepath.Join("admin", pluralName)
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
		"RouteName":  pluralName,
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
	utils.UpdateNavFile(pluralName)

	// Update admin/index.html
	utils.UpdateIndexFile(pluralName)

	fmt.Printf("Admin interface for %s generated in %s\n", singularName, adminDir)
}
