package cmd

import (
	"github.com/ananthakumaran/paisa/internal/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serve the WEB UI",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := gorm.Open(sqlite.Open(viper.GetString("db_path")), &gorm.Config{})
		if err != nil {
			log.Fatal(err)
		}
		server.Listen(db)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
