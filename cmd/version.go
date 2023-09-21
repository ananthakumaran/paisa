package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Version:", "0.5.2")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
