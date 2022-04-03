package cmd

import (
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search mutual fund",
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
