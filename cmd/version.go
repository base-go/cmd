package cmd

import (
	"fmt"

	"github.com/base-go/cmd/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Base CLI",
	Long:  `All software has versions. This is Base's.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.GetBuildInfo())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
