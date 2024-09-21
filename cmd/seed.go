package cmd

import (
	"bytes"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "seed the application",
	Long:  `seed the application by running 'go run main.go' in the current directory.`,
	Run:   seedApplication,
}

func init() {
	rootCmd.AddCommand(seedCmd)
}

func UpdateSeedFile(packageName, structName string) error {
	seedFilePath := "app/seed.go"

	// Read the current content of seed.go
	content, err := os.ReadFile(seedFilePath)
	if err != nil {
		// If the file doesn't exist, create it with the initial structure
		if os.IsNotExist(err) {
			initialContent := `package app

import (
	"base/core/module"
)

func InitializeSeeders() []module.Seeder {
	seeders := []module.Seeder{
		// Add other seeders here
	}
	return seeders
}
`
			err = os.WriteFile(seedFilePath, []byte(initialContent), 0644)
			if err != nil {
				return err
			}
			content = []byte(initialContent)
		} else {
			return err
		}
	}

	// Add import for the new module if it doesn't exist
	importStr := fmt.Sprintf("\"base/app/%s\"", packageName)
	content, importAdded := AddImport(content, importStr)

	// Add seeder to the InitializeSeeders function
	seederStr := fmt.Sprintf("\t\t&%s.%sSeeder{},", packageName, structName)
	content, seederAdded := AddSeeder(content, seederStr)

	// Write the updated content back to seed.go only if changes were made
	if importAdded || seederAdded {
		return os.WriteFile(seedFilePath, content, 0644)
	}

	return nil
}

func AddSeeder(content []byte, seederStr string) ([]byte, bool) {
	// Check if the seeder already exists
	if bytes.Contains(content, []byte(seederStr)) {
		return content, false
	}

	// Find the position of "// Add other seeders here"
	markerIndex := bytes.LastIndex(content, []byte("// Add other seeders here"))
	if markerIndex == -1 {
		return content, false
	}

	// Insert the new seeder before the marker
	updatedContent := append(content[:markerIndex], append([]byte(seederStr+"\n\t\t"), content[markerIndex:]...)...)

	return updatedContent, true
}
