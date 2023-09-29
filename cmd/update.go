package cmd

import (
	"github.com/ananthakumaran/paisa/internal/model"
	"github.com/ananthakumaran/paisa/internal/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var updateJournal bool
var updateCommodities bool
var updatePortfolios bool

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Sync journal data",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := utils.OpenDB()
		if err != nil {
			log.Fatal(err)
		}

		syncAll := !updateJournal && !updateCommodities && !updatePortfolios

		if syncAll || updateJournal {
			message, err := model.SyncJournal(db)
			if err != nil {
				log.Fatal(message)
			}
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
	updateCmd.Flags().BoolVarP(&updatePortfolios, "portfolio", "p", false, "update mutualfund portfolios")
}
