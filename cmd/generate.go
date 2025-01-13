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
	pluralName := utils.ToPlural(singularName)
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

	// Process fields
	fieldStructs := utils.GenerateFieldStructs(fields)

	// Generate files using templates
	templates := []struct {
		targetDir  string
		filename   string
		template   string
	}{
		{
			targetDir:  filepath.Join("app", "models"),
			filename:   fmt.Sprintf("%s.go", dirName),
			template:   "model.tmpl",
		},
		{
			targetDir:  filepath.Join("app", pluralDirName),
			filename:   "service.go",
			template:   "service.tmpl",
		},
		{
			targetDir:  filepath.Join("app", pluralDirName),
			filename:   "controller.go",
			template:   "controller.tmpl",
		},
		{
			targetDir:  filepath.Join("app", pluralDirName),
			filename:   "module.go",
			template:   "module.tmpl",
		},
	}

	for _, tmpl := range templates {
		utils.GenerateFileFromTemplate(
			tmpl.targetDir,
			tmpl.filename,
			tmpl.template,
			structName,
			pluralName,
			packageName,
			fieldStructs,
		)
	}

	// Update init.go
	if err := updateInitGo(pluralDirName, structName); err != nil {
		fmt.Printf("Error updating init.go: %v\n", err)
		return
	}

	fmt.Printf("Generated module %s with fields: %v\n", structName, fields)
}

// updateInitGo adds the module initialization to init.go
func updateInitGo(moduleName, structName string) error {
	initFile := filepath.Join("app", "init.go")

	// Read existing content
	content, err := os.ReadFile(initFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error reading init.go: %v", err)
	}

	// Check if module is already initialized
	if strings.Contains(string(content), fmt.Sprintf("New%sModule", structName)) {
		fmt.Printf("Module initializer for %s already exists in init.go\n", strings.ToLower(structName))
		return nil
	}

	// Create init.go if it doesn't exist
	if os.IsNotExist(err) {
		content = []byte(`package app

import (
	"github.com/gin-gonic/gin"
)

// InitializeModules initializes all modules
func InitializeModules(r *gin.Engine) {
}
`)
	}

	// Find the closing brace of InitializeModules function
	contentStr := string(content)
	insertPos := strings.LastIndex(contentStr, "}")

	if insertPos == -1 {
		return fmt.Errorf("could not find InitializeModules function in init.go")
	}

	// Add import if needed
	importStr := fmt.Sprintf(`"%s/app/%s"`, "base", moduleName)
	if !strings.Contains(contentStr, importStr) {
		importPos := strings.Index(contentStr, ")")
		if importPos == -1 {
			return fmt.Errorf("could not find import section in init.go")
		}
		contentStr = contentStr[:importPos] + "\n\t" + importStr + contentStr[importPos:]
	}

	// Add module initialization
	initStr := fmt.Sprintf("\t%s.New%sModule(r)\n", moduleName, structName)
	contentStr = contentStr[:insertPos] + initStr + contentStr[insertPos:]

	// Write back to file
	return os.WriteFile(initFile, []byte(contentStr), 0644)
}
