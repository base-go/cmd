package main

import (
	"archive/zip"
	"bytes"
	"embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
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

var newCmd = &cobra.Command{
	Use:   "new [project_name]",
	Short: "Create a new project",
	Long:  `Create a new project by cloning the base repository and changing to the new directory.`,
	Args:  cobra.ExactArgs(1),
	Run:   createNewProject,
}

var startCmd = &cobra.Command{
	Use:     "start",
	Aliases: []string{"s"},
	Short:   "Start the application",
	Long:    `Start the application by running 'go run main.go' in the current directory.`,
	Run:     startApplication,
}

var generateCmd = &cobra.Command{
	Use:   "g [name] [field:type...] [--admin]",
	Short: "Generate a new module",
	Long:  `Generate a new module with the specified name and fields. Use --admin flag to generate admin interface.`,
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

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Base to the latest version",
	Long:  `Update Base to the latest version by re-running the installation script.`,
	Run:   updateBase,
}

func init() {
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(destroyCmd)
	rootCmd.AddCommand(updateCmd)
	generateCmd.Flags().Bool("admin", false, "Generate admin interface")
}

func startApplication(cmd *cobra.Command, args []string) {
	// Check if main.go exists in the current directory
	if _, err := os.Stat("main.go"); os.IsNotExist(err) {
		fmt.Println("Error: main.go not found in the current directory.")
		fmt.Println("Make sure you are in the root directory of your Base project.")
		return
	}

	// Run "go run main.go"
	goCmd := exec.Command("go", "run", "main.go")
	goCmd.Stdout = os.Stdout
	goCmd.Stderr = os.Stderr

	fmt.Println("Starting the application...")
	err := goCmd.Run()
	if err != nil {
		fmt.Printf("Error starting the application: %v\n", err)
		return
	}
}

func createNewProject(cmd *cobra.Command, args []string) {
	projectName := args[0]
	archiveURL := "https://github.com/base-go/base/archive/main.zip" // URL to the zip archive of your base project

	// Create the project directory
	err := os.Mkdir(projectName, 0755)
	if err != nil {
		fmt.Printf("Error creating project directory: %v\n", err)
		return
	}

	// Download the archive
	resp, err := http.Get(archiveURL)
	if err != nil {
		fmt.Printf("Error downloading project template: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Create a temporary file to store the zip
	tmpZip, err := os.CreateTemp("", "base-project-*.zip")
	if err != nil {
		fmt.Printf("Error creating temporary file: %v\n", err)
		return
	}
	defer os.Remove(tmpZip.Name())

	// Copy the zip content to the temporary file
	_, err = io.Copy(tmpZip, resp.Body)
	if err != nil {
		fmt.Printf("Error saving project template: %v\n", err)
		return
	}
	tmpZip.Close()

	// Unzip the file
	err = unzip(tmpZip.Name(), projectName)
	if err != nil {
		fmt.Printf("Error extracting project template: %v\n", err)
		return
	}

	// Move contents from the subdirectory to the project root
	files, err := os.ReadDir(filepath.Join(projectName, "base-main"))
	if err != nil {
		fmt.Printf("Error reading template directory: %v\n", err)
		return
	}

	for _, f := range files {
		oldPath := filepath.Join(projectName, "base-main", f.Name())
		newPath := filepath.Join(projectName, f.Name())
		err = os.Rename(oldPath, newPath)
		if err != nil {
			fmt.Printf("Error moving file %s: %v\n", f.Name(), err)
		}
	}

	// Remove the now-empty subdirectory
	os.RemoveAll(filepath.Join(projectName, "base-main"))

	// Get the absolute path of the new project directory
	absPath, err := filepath.Abs(projectName)
	if err != nil {
		fmt.Printf("Error getting absolute path: %v\n", err)
		return
	}

	fmt.Printf("New project '%s' created successfully at %s\n", projectName, absPath)
	fmt.Println("You can now start working on your new project!")
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

func generateModule(cmd *cobra.Command, args []string) {
	singularName := args[0]
	pluralName := toLowerPlural(singularName)
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
	generateFileFromTemplate(filepath.Join("app", "models"), fmt.Sprintf("%s.go", toLower(singularName)), "templates/model.tmpl", singularName, pluralName, fields)
	generateFileFromTemplate(filepath.Join("app", pluralName), "controller.go", "templates/controller.tmpl", singularName, pluralName, fields)
	generateFileFromTemplate(filepath.Join("app", pluralName), "service.go", "templates/service.tmpl", singularName, pluralName, fields)
	generateFileFromTemplate(filepath.Join("app", pluralName), "mod.go", "templates/mod.tmpl", singularName, pluralName, fields)

	// Update app/init.go to register the new module
	if err := updateInitFile(singularName, pluralName); err != nil {
		fmt.Printf("Error updating app/init.go: %v\n", err)
		return
	}

	adminFlag, _ := cmd.Flags().GetBool("admin")
	if adminFlag {
		generateAdminInterface(singularName, pluralName, fields)
	}

	fmt.Printf("Module %s generated successfully with fields: %v\n", singularName, fields)
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

			goType := getGoType(fieldType) // Use only the type part

			if fieldType == "belongs_to" {
				if len(parts) > 2 {
					associatedType = toTitle(parts[2])
					// Add ID field for belongs_to relationships
					fieldStructs = append(fieldStructs, struct {
						Name           string
						Type           string
						JSONName       string
						DBName         string
						AssociatedType string
						PluralType     string
					}{
						Name:     name + "ID",
						Type:     "uint",
						JSONName: jsonName + "Id",
						DBName:   dbName + "_id",
					})
					// Set the type for the association field
					goType = associatedType
				}
			} else if fieldType == "has_many" {
				if len(parts) > 2 {
					associatedType = toTitle(parts[2])
					pluralType = pluralizeClient.Plural(toLower(parts[2]))
					goType = "[]" + associatedType
				}
			} else if fieldType == "has_one" {
				if len(parts) > 2 {
					associatedType = toTitle(parts[2])
					goType = associatedType
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
				Type:           goType,
				JSONName:       jsonName,
				DBName:         dbName,
				AssociatedType: associatedType,
				PluralType:     pluralType,
			})
		}
	}

	return fieldStructs
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
func generateAdminInterface(singularName, pluralName string, fields []string) {
	adminDir := filepath.Join("admin", pluralName)
	if err := os.MkdirAll(adminDir, os.ModePerm); err != nil {
		fmt.Printf("Error creating admin directory %s: %v\n", adminDir, err)
		return
	}

	adminTemplate := "templates/admin_interface.tmpl"

	tmplContent, err := templateFS.ReadFile(adminTemplate)
	if err != nil {
		fmt.Printf("Error reading admin template %s: %v\n", adminTemplate, err)
		return
	}

	funcMap := template.FuncMap{
		"getInputType": getInputType,
	}

	tmpl, err := template.New(filepath.Base(adminTemplate)).Funcs(funcMap).Parse(string(tmplContent))
	if err != nil {
		fmt.Printf("Error parsing admin template: %v\n", err)
		return
	}

	fieldStructs := generateFieldStructs(fields)

	data := map[string]interface{}{
		"StructName": toTitle(singularName),
		"PluralName": toTitle(pluralName),
		"RouteName":  pluralName,
		"Fields":     fieldStructs,
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

	fmt.Printf("Admin interface for %s generated in %s\n", singularName, adminDir)
}

func getInputType(goType string) string {
	switch goType {
	case "int", "int64", "uint", "uint64":
		return "number"
	case "float64":
		return "number"
	case "bool":
		return "checkbox"
	case "time.Time":
		return "datetime-local"
	default:
		return "text"
	}
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
	switch t {
	case "int":
		return "int"
	case "string", "text":
		return "string"
	case "datetime", "time":
		return "time.Time"
	case "float":
		return "float64"
	case "bool":
		return "bool"
	default:
		return t // Return the type as-is for custom types
	}
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

func updateBase(cmd *cobra.Command, args []string) {
	fmt.Println("Updating Base to the latest version...")

	// Define the installation script URL
	scriptURL := "https://raw.githubusercontent.com/base-go/cmd/main/install.sh"

	// Create a temporary file to store the script
	tmpFile, err := os.CreateTemp("", "base-install-*.sh")
	if err != nil {
		fmt.Printf("Error creating temporary file: %v\n", err)
		return
	}
	defer os.Remove(tmpFile.Name())

	// Download the installation script
	downloadCmd := exec.Command("curl", "-fsSL", scriptURL, "-o", tmpFile.Name())
	downloadCmd.Stdout = os.Stdout
	downloadCmd.Stderr = os.Stderr
	if err := downloadCmd.Run(); err != nil {
		fmt.Printf("Error downloading installation script: %v\n", err)
		return
	}

	// Make the script executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		fmt.Printf("Error making script executable: %v\n", err)
		return
	}

	// Run the installation script
	updateCmd := exec.Command("bash", tmpFile.Name())
	updateCmd.Stdout = os.Stdout
	updateCmd.Stderr = os.Stderr
	if err := updateCmd.Run(); err != nil {
		fmt.Printf("Error updating Base: %v\n", err)
		return
	}

	fmt.Println("Base has been successfully updated to the latest version.")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
