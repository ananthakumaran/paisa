package cmd

import (
	"github.com/ananthakumaran/paisa/internal/model"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var updateJournal bool
var updateCommodities bool
var updatePortfolios bool

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Sync journal data",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := gorm.Open(sqlite.Open(viper.GetString("db_path")), &gorm.Config{})
		if err != nil {
			log.Fatal(err)
		}

		syncAll := !updateJournal && !updateCommodities && !updatePortfolios

		if syncAll || updateJournal {
			model.SyncJournal(db)
		}

		if syncAll || updateCommodities {
			model.SyncCommodities(db)
		}

		if syncAll || updatePortfolios {
			model.SyncPortfolios(db)
		}

		if syncAll {
			model.SyncCII(db)
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolVarP(&updateJournal, "journal", "j", false, "update journal")
	updateCmd.Flags().BoolVarP(&updateCommodities, "commodity", "c", false, "update commodities")
	updateCmd.Flags().BoolVarP(&updatePortfolios, "portfolio", "p", false, "update portfolios")
}
