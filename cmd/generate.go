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
		targetDir    string
		filename     string
		templateFile string
	}{
		{
			targetDir:    filepath.Join("app", "models"),
			filename:     fmt.Sprintf("%s.go", dirName),
			templateFile: "templates/model.tmpl",
		},
		{
			targetDir:    filepath.Join("app", pluralDirName),
			filename:     "service.go",
			templateFile: "templates/service.tmpl",
		},
		{
			targetDir:    filepath.Join("app", pluralDirName),
			filename:     "controller.go",
			templateFile: "templates/controller.tmpl",
		},
		{
			targetDir:    filepath.Join("app", pluralDirName),
			filename:     "module.go",
			templateFile: "templates/module.tmpl",
		},
	}

	for _, tmpl := range templates {
		utils.GenerateFileFromTemplate(
			tmpl.targetDir,
			tmpl.filename,
			tmpl.templateFile,
			structName,
			pluralName,
			packageName,
			fieldStructs,
		)
	}

	// Update init.go
	initFile := filepath.Join("app", "init.go")
	if _, err := os.Stat(initFile); err == nil {
		content, err := os.ReadFile(initFile)
		if err != nil {
			fmt.Printf("Error reading init.go: %v\n", err)
			return
		}

		// Create the module initializer line
		moduleInitializer := fmt.Sprintf(`"%s": func(db *gorm.DB, router *gin.RouterGroup, log logger.Logger, emitter *emitter.Emitter, activeStorage *storage.ActiveStorage) module.Module { 
            return %s.New%sModule(db, router, log, emitter, activeStorage)
        },`,
			packageName, packageName, structName)

		// Check if the module initializer already exists
		if strings.Contains(string(content), fmt.Sprintf(`"%s":`, packageName)) {
			fmt.Printf("Module initializer for %s already exists in init.go\n", packageName)
			return
		}

		// Insert the initializer before the marker comment
		markerComment := "// MODULE_INITIALIZER_MARKER"
		contentStr := string(content)
		markerIndex := strings.Index(contentStr, markerComment)
		if markerIndex == -1 {
			fmt.Println("Could not find marker comment in init.go")
			return
		}

		// Find the start of the line containing the marker
		lineStart := strings.LastIndex(contentStr[:markerIndex], "\n") + 1
		newContent := contentStr[:lineStart] + "\t\t" + moduleInitializer + "\n\t\t" + contentStr[lineStart:]

		if err := os.WriteFile(initFile, []byte(newContent), 0644); err != nil {
			fmt.Printf("Error writing to init.go: %v\n", err)
			return
		}
	}

	fmt.Printf("Successfully generated %s module\n", structName)
}
