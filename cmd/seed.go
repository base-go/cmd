package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var seedCmd = &cobra.Command{
	Use:     "seed [seedName]",
	Aliases: []string{"s"},
	Short:   "Run database seeds",
	Long:    `Run database seeds. Use 'all' to run all seeders or specify a seeder name.`,
	Run:     runSeed,
}

func init() {
	rootCmd.AddCommand(seedCmd)
}

func runSeed(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Seed name is required. Use 'all' to run all seeders or specify a seeder name.")
		return
	}

	// Ensure we have the necessary environment variables
	requiredEnvVars := []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME"}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			fmt.Printf("Error: %s environment variable is not set.\n", envVar)
			return
		}
	}

	// Run the internal seed command
	internalCmd := exec.Command(os.Args[0], append([]string{"internal-seed"}, args...)...)
	internalCmd.Stdout = os.Stdout
	internalCmd.Stderr = os.Stderr
	internalCmd.Env = os.Environ()

	err := internalCmd.Run()
	if err != nil {
		fmt.Printf("Error running seeds: %v\n", err)
		return
	}
}
