package cmd

import (
	"os"

	"github.com/ananthakumaran/paisa/internal/generator"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generates a sample config and journal file",
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		generator.Demo(cwd)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
