package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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

	fmt.Printf("Successfully generated %s module\n", naming.Model)
}
