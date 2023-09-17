package cmd

import (
	"os"

	"github.com/ananthakumaran/paisa/internal/model"
	"github.com/ananthakumaran/paisa/internal/server"
	"github.com/ananthakumaran/paisa/internal/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var port int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serve the WEB UI",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := utils.OpenDB()
		model.AutoMigrate(db)

		if os.Getenv("PAISA_DEBUG") == "true" {
			db = db.Debug()
		}

		if err != nil {
			log.Fatal(err)
		}
		server.Listen(db, port)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&port, "port", "p", 7500, "port to listen on")
}
