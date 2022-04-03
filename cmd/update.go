package cmd

import (
	"github.com/ananthakumaran/paisa/internal/model"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Sync journal data",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := gorm.Open(sqlite.Open(viper.GetString("db_path")), &gorm.Config{})
		if err != nil {
			log.Fatal(err)
		}
		model.Sync(db)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
