package main

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gertd/go-pluralize"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var pluralizeClient *pluralize.Client

func init() {
	pluralizeClient = pluralize.NewClient()
}

//go:embed templates/*
var templateFS embed.FS

var rootCmd = &cobra.Command{
	Use:   "base [command] [args]",
	Short: "Generate or destroy modules for the application",
	Long:  `A command-line tool to generate new modules with predefined structure or destroy existing modules for the application.`,
}

var generateCmd = &cobra.Command{
	Use:   "g [name] [field:type...]",
	Short: "Generate a new module",
	Long:  `Generate a new module with the specified name and fields.`,
	Args:  cobra.MinimumNArgs(1),
	Run:   generateModule,
}

var destroyCmd = &cobra.Command{
	Use:   "d [name]",
	Short: "Destroy an existing module",
	Long:  `Destroy an existing module with the specified name.`,
	Args:  cobra.ExactArgs(1),
	Run:   destroyModule,
}

func init() {
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(destroyCmd)
}

func generateModule(cmd *cobra.Command, args []string) {
	singularName := args[0]
	pluralName := toLowerPlural(singularName)
	fields := args[1:]

	// Create module directory (lowercase plural)
	moduleDir := filepath.Join("app", pluralName)
	err := os.MkdirAll(moduleDir, os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating directory %s: %v\n", moduleDir, err)
		return
	}

	// Generate files using templates
	generateFileFromTemplate(moduleDir, "model.go", "templates/model.tmpl", singularName, pluralName, fields)
	generateFileFromTemplate(moduleDir, "controller.go", "templates/controller.tmpl", singularName, pluralName, fields)
	generateFileFromTemplate(moduleDir, "service.go", "templates/service.tmpl", singularName, pluralName, fields)
	generateFileFromTemplate(moduleDir, "mod.go", "templates/mod.tmpl", singularName, pluralName, fields)

	// Update app/init.go to register the new module
	err = updateInitFile(singularName, pluralName)
	if err != nil {
		fmt.Printf("Error updating app/init.go: %v\n", err)
		return
	}

	fmt.Printf("Generating module %s with fields: %v\n", args[0], fields)
	fmt.Printf("Module '%s' generated successfully in directory '%s'.\n", singularName, moduleDir)
	fmt.Println("Module registered in app/init.go successfully!")
}

func updateInitFile(singularName, pluralName string) error {
	initFilePath := "app/init.go"

	// Read the current content of init.go
	content, err := os.ReadFile(initFilePath)
	if err != nil {
		return err
	}

	// Add import for the new module if it doesn't exist
	importStr := fmt.Sprintf("\"base/app/%s\"", pluralName)
	content, importAdded := addImport(content, importStr)

	// Add module initializer if it doesn't exist
	content, initializerAdded := addModuleInitializer(content, pluralName, singularName)

	// Write the updated content back to init.go only if changes were made
	if importAdded || initializerAdded {
		return os.WriteFile(initFilePath, content, 0644)
	}

	return nil
}

func addImport(content []byte, importStr string) ([]byte, bool) {
	// Check if the import already exists
	if bytes.Contains(content, []byte(importStr)) {
		return content, false
	}

	// Find the position of "import ("
	importPos := bytes.Index(content, []byte("import ("))
	if importPos == -1 {
		// If "import (" is not found, return original content
		return content, false
	}

	// Position to insert the new import (after "import (" and newline)
	insertPos := importPos + len("import (") + 1

	// Create the new import line with proper indentation
	newImportLine := []byte("\t" + importStr + "\n")

	// Insert the new import line
	updatedContent := append(content[:insertPos], append(newImportLine, content[insertPos:]...)...)

	return updatedContent, true
}

func addModuleInitializer(content []byte, pluralName, singularName string) ([]byte, bool) {
	contentStr := string(content)

	// Find the module initializer marker
	markerIndex := strings.Index(contentStr, "// MODULE_INITIALIZER_MARKER")
	if markerIndex == -1 {
		return content, false
	}

	// Check if the module already exists
	if strings.Contains(contentStr[:markerIndex], fmt.Sprintf(`"%s":`, pluralName)) {
		return content, false
	}

	// Create the new initializer
	newInitializer := fmt.Sprintf(`        "%s": func(db *gorm.DB, router *gin.RouterGroup) module.Module { return %s.New%sModule(db, router) },`,
		pluralName, pluralName, toTitle(singularName))

	// Insert the new initializer before the marker
	updatedContent := contentStr[:markerIndex] + newInitializer + "\n        " + contentStr[markerIndex:]

	return []byte(updatedContent), true
}

func generateFileFromTemplate(dir, filename, templateFile, singularName, pluralName string, fields []string) {
	tmplContent, err := templateFS.ReadFile(templateFile)
	if err != nil {
		fmt.Printf("Error reading template %s: %v\n", templateFile, err)
		return
	}

	funcMap := template.FuncMap{
		"toLower": strings.ToLower,
	}

	tmpl, err := template.New(filepath.Base(templateFile)).Funcs(funcMap).Parse(string(tmplContent))
	if err != nil {
		fmt.Printf("Error parsing template %s: %v\n", templateFile, err)
		return
	}

	fieldStructs := generateFieldStructs(fields)

	data := map[string]interface{}{
		"PackageName": pluralName,
		"StructName":  toTitle(singularName),
		"PluralName":  toTitle(pluralName),
		"RouteName":   pluralName,
		"Fields":      fieldStructs,
		"TableName":   pluralName,
	}

	filePath := filepath.Join(dir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", filePath, err)
		return
	}
	defer file.Close()

	err = tmpl.Execute(file, data)
	if err != nil {
		fmt.Printf("Error executing template for %s: %v\n", filename, err)
	}
}

func generateFieldStructs(fields []string) []struct {
	Name           string
	Type           string
	JSONName       string
	DBName         string
	AssociatedType string
	PluralType     string
} {
	var fieldStructs []struct {
		Name           string
		Type           string
		JSONName       string
		DBName         string
		AssociatedType string
		PluralType     string
	}

	for _, field := range fields {
		parts := strings.Split(field, ":")
		if len(parts) >= 2 {
			name := toTitle(parts[0])
			fieldType := parts[1]
			jsonName := toLower(parts[0])
			dbName := toLower(parts[0])
			var associatedType, pluralType string

			if fieldType == "belongs_to" || fieldType == "has_many" || fieldType == "has_one" {
				if len(parts) >= 3 {
					associatedType = toTitle(parts[2])
					pluralType = pluralizeClient.Plural(toLower(parts[2]))
				} else {
					associatedType = "interface{}"
					pluralType = "interfaces"
				}
			}

			fieldStructs = append(fieldStructs, struct {
				Name           string
				Type           string
				JSONName       string
				DBName         string
				AssociatedType string
				PluralType     string
			}{
				Name:           name,
				Type:           fieldType,
				JSONName:       jsonName,
				DBName:         dbName,
				AssociatedType: associatedType,
				PluralType:     pluralType,
			})
		}
	}

	// Debug logging
	for _, field := range fieldStructs {
		fmt.Printf("Field: Name=%s, Type=%s, JSONName=%s, DBName=%s, AssociatedType=%s, PluralType=%s\n",
			field.Name, field.Type, field.JSONName, field.DBName, field.AssociatedType, field.PluralType)
	}

	return fieldStructs
}

// destroyModule destroys an existing module
func destroyModule(cmd *cobra.Command, args []string) {
	singularName := args[0]

	// Check if the module exists
	_, err := os.Stat(filepath.Join("app", toLowerPlural(singularName)))
	if os.IsNotExist(err) {
		fmt.Printf("Module '%s' does not exist.\n", singularName)
		return
	}

	pluralName := toLowerPlural(singularName)

	// Delete module directory
	moduleDir := filepath.Join("app", pluralName)
	err = os.RemoveAll(moduleDir)
	if err != nil {
		fmt.Printf("Error removing directory %s: %v\n", moduleDir, err)
		return
	}

	// Update app/init.go to unregister the module
	err = updateInitFileForDestroy(pluralName)
	if err != nil {
		fmt.Printf("Error updating app/init.go: %v\n", err)
		return
	}

	fmt.Printf("Module '%s' destroyed successfully.\n", singularName)
	fmt.Println("Module unregistered from app/init.go successfully!")
}

func updateInitFileForDestroy(pluralName string) error {
	initFilePath := "app/init.go"

	// Read the current content of init.go
	content, err := os.ReadFile(initFilePath)
	if err != nil {
		return err
	}

	// Remove import for the module
	importStr := fmt.Sprintf("\"base/app/%s\"", pluralName)
	content = removeImport(content, importStr)

	// Remove module initializer
	content = removeModuleInitializer(content, pluralName)

	// Write the updated content back to init.go
	return os.WriteFile(initFilePath, content, 0644)
}

func removeImport(content []byte, importStr string) []byte {
	lines := bytes.Split(content, []byte("\n"))
	var newLines [][]byte

	for _, line := range lines {
		if !bytes.Contains(line, []byte(importStr)) {
			newLines = append(newLines, line)
		}
	}

	return bytes.Join(newLines, []byte("\n"))
}

func removeModuleInitializer(content []byte, pluralName string) []byte {
	lines := bytes.Split(content, []byte("\n"))
	var newLines [][]byte

	for _, line := range lines {
		if !bytes.Contains(line, []byte(fmt.Sprintf(`"%s":`, pluralName))) {
			newLines = append(newLines, line)
		}
	}

	return bytes.Join(newLines, []byte("\n"))
}

func getGoType(t string) string {
	parts := strings.Split(t, ":")
	baseType := parts[0]

	goType := ""
	switch baseType {
	case "int":
		goType = "int"
	case "string", "text":
		goType = "string"
	case "datetime", "time":
		goType = "time.Time"
	case "float":
		goType = "float64"
	case "bool":
		goType = "bool"
	case "belongs_to":
		if len(parts) > 1 {
			goType = "*" + toTitle(parts[1]) // Pointer to the associated type
		} else {
			goType = "*interface{}" // Generic pointer if no specific type is provided
		}
	case "has_many":
		if len(parts) > 1 {
			goType = "[]" + toTitle(parts[1]) // Slice of the associated type
		} else {
			goType = "[]interface{}" // Generic slice if no specific type is provided
		}
	case "has_one":
		if len(parts) > 1 {
			goType = "*" + toTitle(parts[1]) // Pointer to the associated type
		} else {
			goType = "*interface{}" // Generic pointer if no specific type is provided
		}
	default:
		fmt.Printf("Warning: Unexpected field type '%s'. Defaulting to string.\n", baseType)
		goType = "string" // Default to string for unknown types
	}

	fmt.Printf("Field type '%s' mapped to Go type '%s'\n", t, goType)
	return goType
}

func toLower(s string) string {
	return strings.ToLower(s)
}

func toTitle(s string) string {
	return cases.Title(language.Und).String(s)
}

func toLowerPlural(s string) string {
	return strings.ToLower(pluralizeClient.Plural(s))
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
