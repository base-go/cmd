package cmd

import (
	"fmt"
	"os"
	"path/filepath"

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
	pluralName := utils.ToLowerPlural(singularName)
	fields := args[1:]

	// Create directories
	dirs := []string{
		filepath.Join("app", "models"),
		filepath.Join("app", pluralName),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			return
		}
	}

	// Generate files using templates
	utils.GenerateFileFromTemplate(
		filepath.Join("app", "models"),
		fmt.Sprintf("%s.go", utils.ToLower(singularName)),
		"templates/model.tmpl",
		singularName,
		pluralName,
		fields,
	)
	utils.GenerateFileFromTemplate(
		filepath.Join("app", pluralName),
		"controller.go",
		"templates/controller.tmpl",
		singularName,
		pluralName,
		fields,
	)
	utils.GenerateFileFromTemplate(
		filepath.Join("app", pluralName),
		"service.go",
		"templates/service.tmpl",
		singularName,
		pluralName,
		fields,
	)
	utils.GenerateFileFromTemplate(
		filepath.Join("app", pluralName),
		"mod.go",
		"templates/mod.tmpl",
		singularName,
		pluralName,
		fields,
	)

	// Update app/init.go to register the new module
	if err := utils.UpdateInitFile(singularName, pluralName); err != nil {
		fmt.Printf("Error updating app/init.go: %v\n", err)
		return
	}

	adminFlag, _ := cmd.Flags().GetBool("admin")
	if adminFlag {
		generateAdminInterface(singularName, pluralName, fields)
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
