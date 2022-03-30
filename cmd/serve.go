package cmd

import (
	"github.com/ananthakumaran/paisa/internal/server"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serve the WEB UI",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		server.Listen()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
