package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gorm.io/driver/mysql"
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

	// Initialize the database connection
	db, err := initDB()
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
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

func initDB() (*gorm.DB, error) {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
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
	// Here, instead of compiling and loading a plugin, you would directly call
	// the seeding function. You'll need to implement a way to map seed names
	// to their respective functions.
	seeder, err := getSeedFunction(seedName)
	if err != nil {
		return fmt.Errorf("error getting seed function: %w", err)
	}

	if err := seeder(db); err != nil {
		return fmt.Errorf("error running seeder: %w", err)
	}

	fmt.Printf("Seeded %s successfully\n", seedName)
	return nil
}

// Implement this function to return the appropriate seeding function
func getSeedFunction(seedName string) (func(*gorm.DB) error, error) {
	// This is where you'll map seed names to their respective functions
	// For example:
	// switch seedName {
	// case "users":
	//     return seedUsers, nil
	// case "posts":
	//     return seedPosts, nil
	// default:
	//     return nil, fmt.Errorf("unknown seed: %s", seedName)
	// }
	return nil, fmt.Errorf("getSeedFunction not implemented")
}
