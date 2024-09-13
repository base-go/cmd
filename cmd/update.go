package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Base to the latest version",
	Long:  `Update Base to the latest version by re-running the installation script.`,
	Run:   updateBase,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func updateBase(cmd *cobra.Command, args []string) {
	fmt.Println("Updating Base to the latest version...")

	// Define the installation script URL
	scriptURL := "https://raw.githubusercontent.com/base-go/cmd/main/install.sh"

	// Create a temporary file to store the script
	tmpFile, err := os.CreateTemp("", "base-install-*.sh")
	if err != nil {
		fmt.Printf("Error creating temporary file: %v\n", err)
		return
	}
	defer os.Remove(tmpFile.Name())

	// Download the installation script
	downloadCmd := exec.Command("curl", "-fsSL", scriptURL, "-o", tmpFile.Name())
	downloadCmd.Stdout = os.Stdout
	downloadCmd.Stderr = os.Stderr
	if err := downloadCmd.Run(); err != nil {
		fmt.Printf("Error downloading installation script: %v\n", err)
		return
	}

	// Make the script executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		fmt.Printf("Error making script executable: %v\n", err)
		return
	}

	// Run the installation script
	runCmd := exec.Command("bash", tmpFile.Name())
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr
	if err := runCmd.Run(); err != nil {
		fmt.Printf("Error updating Base: %v\n", err)
		return
	}

	fmt.Println("Base has been successfully updated to the latest version.")
}
