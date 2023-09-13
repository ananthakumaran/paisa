package cmd

import (
	"strings"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model"
	"github.com/ananthakumaran/paisa/internal/model/mutualfund/scheme"
	"github.com/ananthakumaran/paisa/internal/scraper/mutualfund"
	"github.com/logrusorgru/aurora"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var update bool

var mutualfundCmd = &cobra.Command{
	Use:   "mutualfund",
	Short: "Search mutual fund",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := gorm.Open(sqlite.Open(config.GetConfig().DBPath), &gorm.Config{})
		if err != nil {
			log.Fatal(err)
		}

		model.AutoMigrate(db)
		count := scheme.Count(db)
		if update || count == 0 {
			schemes, err := mutualfund.GetSchemes()
			if err != nil {
				log.Fatal(err)
			}
			scheme.UpsertAll(db, schemes)
		} else {
			log.Info("Using cached results; pass '-u' to update the cache")
		}
		amc := promptAMC(db)
		name := promptName(db, amc)
		scheme := scheme.FindScheme(db, amc, name)
		log.Info("Mutual Fund Scheme Code: ", aurora.Bold(scheme.Code))
	},
}

func init() {
	searchCmd.AddCommand(mutualfundCmd)
	mutualfundCmd.Flags().BoolVarP(&update, "update", "u", false, "update the Mutual Fund Scheme list")
}

func promptAMC(db *gorm.DB) string {
	amcs := scheme.GetAMCs(db)
	return prompt("AMC", amcs)
}

func promptName(db *gorm.DB, amc string) string {
	names := scheme.GetNAVNames(db, amc)
	return prompt("Fund Name", names)
}

func prompt(label string, list []string) string {
	searcher := func(input string, index int) bool {
		item := list[index]
		item = strings.Replace(strings.ToLower(item), " ", "", -1)
		words := strings.Split(strings.ToLower(input), " ")
		for _, word := range words {
			if strings.TrimSpace(word) != "" && !strings.Contains(item, word) {
				return false
			}
		}
		return true
	}

	prompt := promptui.Select{
		Label:             label,
		Items:             list,
		Size:              10,
		Searcher:          searcher,
		StartInSearchMode: true,
	}

	_, item, err := prompt.Run()

	if err != nil {
		log.Fatal(err)
	}
	return item
}
