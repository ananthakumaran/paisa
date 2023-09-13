package cmd

import (
	"strings"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model"
	"github.com/ananthakumaran/paisa/internal/model/nps/scheme"
	"github.com/ananthakumaran/paisa/internal/scraper/nps"
	"github.com/logrusorgru/aurora"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var npsUpdate bool

var npsCmd = &cobra.Command{
	Use:   "nps",
	Short: "Search nps fund",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := gorm.Open(sqlite.Open(config.GetConfig().DBPath), &gorm.Config{})
		if err != nil {
			log.Fatal(err)
		}

		model.AutoMigrate(db)
		count := scheme.Count(db)
		if update || count == 0 {
			schemes, err := nps.GetSchemes()
			if err != nil {
				log.Fatal(err)
			}
			scheme.UpsertAll(db, schemes)
		} else {
			log.Info("Using cached results; pass '-u' to update the cache")
		}
		pfm := promptPFM(db)
		name := npsPromptName(db, pfm)
		scheme := scheme.FindScheme(db, pfm, name)
		log.Info("NPS Fund Scheme Code: ", aurora.Bold(scheme.SchemeID))
	},
}

func init() {
	searchCmd.AddCommand(npsCmd)
	npsCmd.Flags().BoolVarP(&update, "update", "u", false, "update the NPS Fund Scheme list")
}

func promptPFM(db *gorm.DB) string {
	pfms := scheme.GetPFMs(db)
	return npsPrompt("Pension Fund Manager", pfms)
}

func npsPromptName(db *gorm.DB, pfm string) string {
	names := scheme.GetSchemeNames(db, pfm)
	return npsPrompt("Fund Name", names)
}

func npsPrompt(label string, list []string) string {
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
