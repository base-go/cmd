package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"

	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var internalSeedCmd = &cobra.Command{
	Use:    "internal-seed [seedName]",
	Hidden: true,
	Run:    runInternalSeed,
}

func init() {
	rootCmd.AddCommand(internalSeedCmd)
}

func runInternalSeed(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Seed name is required.")
		return
	}

	seedName := args[0]

	// Try to load the database initialization plugin
	db, err := loadDBPlugin()
	if err != nil {
		fmt.Printf("Error loading database plugin: %v\n", err)
		return
	}

	// Run seeds
	err = runSeeds(db, seedName)
	if err != nil {
		fmt.Printf("Error running seeds: %v\n", err)
		return
	}

	fmt.Println("Seeds executed successfully.")
}

func loadDBPlugin() (*gorm.DB, error) {
	// Build the plugin
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", "core/db_plugin.so", "core/db_plugin.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to build db plugin: %w", err)
	}

	// Load the plugin
	p, err := plugin.Open("core/db_plugin.so")
	if err != nil {
		return nil, fmt.Errorf("failed to open db plugin: %w", err)
	}

	// Look up the InitDB symbol
	initDBSymbol, err := p.Lookup("InitDB")
	if err != nil {
		return nil, fmt.Errorf("failed to find InitDB symbol: %w", err)
	}

	// Assert that initDBSymbol is of the correct type
	initDB, ok := initDBSymbol.(func() (*gorm.DB, error))
	if !ok {
		return nil, fmt.Errorf("unexpected type for InitDB")
	}

	// Call InitDB to get the database connection
	db, err := initDB()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return db, nil
}

func runSeeds(db *gorm.DB, seedName string) error {
	seedsDir := "app/seeds"

	if seedName == "all" {
		return seedAllFiles(db, seedsDir)
	}

	return seedSingleFile(db, seedsDir, seedName)
}

func seedAllFiles(db *gorm.DB, seedsDir string) error {
	files, err := os.ReadDir(seedsDir)
	if err != nil {
		return fmt.Errorf("error reading seeds directory: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".go" && file.Name() != "seed.go" {
			seedName := file.Name()[:len(file.Name())-3] // remove .go extension
			if err := seedSingleFile(db, seedsDir, seedName); err != nil {
				return err
			}
		}
	}

	return nil
}

func seedSingleFile(db *gorm.DB, seedsDir, seedName string) error {
	pluginPath := filepath.Join(seedsDir, seedName+".so")

	// Compile the Go file to a shared object
	if err := compilePlugin(seedsDir, seedName); err != nil {
		return fmt.Errorf("error compiling plugin: %w", err)
	}

	p, err := plugin.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("error opening plugin: %w", err)
	}

	symSeeder, err := p.Lookup("Seeder")
	if err != nil {
		return fmt.Errorf("error looking up Seeder symbol: %w", err)
	}

	seeder, ok := symSeeder.(interface {
		Seed(db *gorm.DB) error
	})
	if !ok {
		return fmt.Errorf("unexpected type from module symbol")
	}

	if err := seeder.Seed(db); err != nil {
		return fmt.Errorf("error running seeder: %w", err)
	}

	fmt.Printf("Seeded %s successfully\n", seedName)
	return nil
}

func compilePlugin(seedsDir, seedName string) error {
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o",
		filepath.Join(seedsDir, seedName+".so"),
		filepath.Join(seedsDir, seedName+".go"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
