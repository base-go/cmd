package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BaseTechStack/basecmd/utils"
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

	// Create naming convention from the input name
	naming := utils.NewNamingConvention(singularName)

	// Create directories (plural names in snake_case)
	dirs := []string{
		filepath.Join("app", "models"),
		filepath.Join("app", naming.DirName),
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
		naming.ModelSnake+".go",
		"model.tmpl",
		naming,
		fieldStructs,
	)

	// Generate service
	utils.GenerateFileFromTemplate(
		filepath.Join("app", naming.DirName),
		"service.go",
		"service.tmpl",
		naming,
		fieldStructs,
	)

	// Generate controller
	utils.GenerateFileFromTemplate(
		filepath.Join("app", naming.DirName),
		"controller.go",
		"controller.tmpl",
		naming,
		fieldStructs,
	)

	// Generate module
	utils.GenerateFileFromTemplate(
		filepath.Join("app", naming.DirName),
		"module.go",
		"module.tmpl",
		naming,
		fieldStructs,
	)

	// Generate validator
	utils.GenerateFileFromTemplate(
		filepath.Join("app", naming.DirName),
		"validator.go",
		"validator.tmpl",
		naming,
		fieldStructs,
	)

	// Generate tests
	if err := utils.GenerateTests(naming, fieldStructs); err != nil {
		fmt.Printf("Error generating tests: %v\n", err)
		return
	}

	// Check if goimports is installed
	if _, err := exec.LookPath("goimports"); err != nil {
		fmt.Println("goimports not found, installing...")
		if err := exec.Command("go", "install", "golang.org/x/tools/cmd/goimports@latest").Run(); err != nil {
			fmt.Printf("Error installing goimports: %v\n", err)
			fmt.Println("Please install goimports manually: go install golang.org/x/tools/cmd/goimports@latest")
			return
		}
		fmt.Println("goimports installed successfully")
	}

	// Run goimports on generated files
	generatedPath := filepath.Join("app", naming.DirName)

	fmt.Println("Running goimports on generated files...")
	// Run goimports on the generated directory
	if err := exec.Command("find", generatedPath, "-name", "*.go", "-exec", "goimports", "-w", "{}", ";").Run(); err != nil {
		fmt.Printf("Error running goimports on %s: %v\n", generatedPath, err)
	}

	// Run goimports on the model file
	modelPath := filepath.Join("app", "models", naming.ModelSnake+".go")
	if err := exec.Command("goimports", "-w", modelPath).Run(); err != nil {
		fmt.Printf("Error running goimports on %s: %v\n", modelPath, err)
	}

	// Automatically add import to main.go if it exists
	if err := addImportToMainGo(naming.DirName); err != nil {
		fmt.Printf("Warning: Could not automatically add import to main.go: %v\n", err)
		fmt.Printf("Please manually add: _ \"base/app/%s\" to your imports in main.go\n", naming.DirName)
	} else {
		fmt.Printf("✅ Automatically added import to main.go\n")
	}

	fmt.Printf("Successfully generated %s module\n", naming.Model)
}

// addImportToMainGo automatically adds the import for the generated module to main.go
func addImportToMainGo(moduleName string) error {
	mainGoPath := "main.go"
	
	// Check if main.go exists
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		return fmt.Errorf("main.go not found in current directory")
	}

	// Read main.go content
	content, err := os.ReadFile(mainGoPath)
	if err != nil {
		return fmt.Errorf("failed to read main.go: %w", err)
	}

	contentStr := string(content)
	importLine := fmt.Sprintf("\t_ \"base/app/%s\"", moduleName)

	// Check if import already exists
	if strings.Contains(contentStr, importLine) {
		return nil // Already imported
	}

	// Find the import section
	importStart := strings.Index(contentStr, "import (")
	if importStart == -1 {
		return fmt.Errorf("could not find import section in main.go")
	}

	// Find the end of the import section
	importEnd := strings.Index(contentStr[importStart:], ")")
	if importEnd == -1 {
		return fmt.Errorf("could not find end of import section in main.go")
	}
	importEnd += importStart

	// Look for the generated modules comment
	generatedComment := "// Import generated modules to trigger their init() functions"
	commentIndex := strings.Index(contentStr, generatedComment)

	if commentIndex == -1 {
		// Add the comment and the import before the closing )
		insertPoint := importEnd
		newContent := contentStr[:insertPoint] + "\n\n\t" + generatedComment + "\n" + importLine + "\n" + contentStr[insertPoint:]
		contentStr = newContent
	} else {
		// Find the last import line after the comment
		commentEnd := commentIndex + len(generatedComment)
		nextImportEnd := importEnd
		
		// Find where to insert (after the last generated import)
		lines := strings.Split(contentStr[commentEnd:nextImportEnd], "\n")
		var lastImportLineIndex int
		for i, line := range lines {
			if strings.Contains(line, "_ \"base/app/") {
				lastImportLineIndex = i
			}
		}
		
		// Insert after the last import line or after the comment if no imports exist
		if lastImportLineIndex > 0 {
			// Insert after the last import
			beforeComment := contentStr[:commentEnd]
			afterComment := contentStr[commentEnd:]
			
			lines = strings.Split(afterComment[:nextImportEnd-commentEnd], "\n")
			lines = append(lines[:lastImportLineIndex+1], append([]string{importLine}, lines[lastImportLineIndex+1:]...)...)
			
			newAfterComment := strings.Join(lines, "\n") + contentStr[nextImportEnd:]
			contentStr = beforeComment + newAfterComment
		} else {
			// Insert right after the comment
			insertPoint := commentEnd + 1
			newContent := contentStr[:insertPoint] + importLine + "\n" + contentStr[insertPoint:]
			contentStr = newContent
		}
	}

	// Write back to file
	if err := os.WriteFile(mainGoPath, []byte(contentStr), 0644); err != nil {
		return fmt.Errorf("failed to write main.go: %w", err)
	}

	return nil
}
