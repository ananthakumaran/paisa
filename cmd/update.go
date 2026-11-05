package cmd

import (
	"errors"

	"github.com/ananthakumaran/paisa/internal/model"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/spf13/cobra"
)

var updateJournal bool
var updateCommodities bool
var updatePortfolios bool

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Sync journal data",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := utils.OpenDB()
		if err != nil {
			return err
		}

		syncAll := !updateJournal && !updateCommodities && !updatePortfolios

		if syncAll || updateJournal {
			message, err := model.SyncJournal(db)
			if err != nil {
				return errors.New(message)
			}
		}

		if syncAll || updateCommodities {
			if _, err := model.SyncCommodities(db); err != nil {
				return err
			}
		}

		if syncAll || updatePortfolios {
			if err := model.SyncPortfolios(db); err != nil {
				return err
			}
		}

		if syncAll {
			if err := model.SyncCII(db); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolVarP(&updateJournal, "journal", "j", false, "update journal")
	updateCmd.Flags().BoolVarP(&updateCommodities, "commodity", "c", false, "update commodities")
	updateCmd.Flags().BoolVarP(&updatePortfolios, "portfolio", "p", false, "update mutualfund portfolios")
}
