package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var feedCmd = &cobra.Command{
	Use:   "feed [mysql_table:json_path] [field_mappings...]",
	Short: "Feed data from JSON file into MySQL table",
	Long:  `Feed data from a JSON file into a specified MySQL table, with optional field mappings.`,
	Args:  cobra.MinimumNArgs(1),
	Run:   feedData,
}

func init() {
	rootCmd.AddCommand(feedCmd)
}

func feedData(cmd *cobra.Command, args []string) {
	// Construct the argument string
	argString := strings.Join(args, " ")

	// Run the feed operation
	runWithArgument("feed " + argString)
}

func runWithArgument(argument string) {
	// Check if main.go exists in the current directory
	if _, err := os.Stat("main.go"); os.IsNotExist(err) {
		fmt.Println("Error: main.go not found in the current directory.")
		fmt.Println("Make sure you are in the root directory of your Base project.")
		return
	}

	// Split the argument string into separate arguments
	args := append([]string{"run", "main.go"}, strings.Split(argument, " ")...)

	// Run "go run main.go" with the given arguments
	goCmd := exec.Command("go", args...)
	goCmd.Stdout = os.Stdout
	goCmd.Stderr = os.Stderr

	fmt.Printf("Running %s operation...\n", strings.Split(argument, " ")[0])
	err := goCmd.Run()
	if err != nil {
		fmt.Printf("Error running %s operation: %v\n", strings.Split(argument, " ")[0], err)
		return
	}
}
