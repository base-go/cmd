package utils

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed "templates/air.tmpl"
var airTemplate string

// GenerateAirFileFromTemplate generates the air.toml configuration file from template
func GenerateAirFileFromTemplate(dir string) error {
	filename := ".air.toml"
	targetPath := filepath.Join(dir, filename)

	// Create the file
	f, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", targetPath, err)
	}
	defer f.Close()

	// Write the template content to the file
	_, err = f.WriteString(airTemplate)
	if err != nil {
		return fmt.Errorf("failed to write template content: %w", err)
	}

	return nil
}
