package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"time"

	"strings"

	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/scraper/mutualfund"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/google/btree"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"math/rand"
)

const START_YEAR = 2015

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "generates a sample config and journal file",
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		generateConfigFile(cwd)
		generateJournalFile(cwd)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func generateConfigFile(cwd string) {
	configFilePath := filepath.Join(cwd, "paisa.yaml")
	config := `
journal_path: '%s'
db_path: '%s'
allocation_targets:
  - name: Debt
    target: 40
    accounts:
      - Assets:Debt:*
  - name: Equity
    target: 60
    accounts:
      - Assets:Equity:*
commodities:
  - name: NIFTY
    type: mutualfund
    code: 120716
  - name: NIFTY_JR
    type: mutualfund
    code: 120684
  - name: ABCBF
    type: mutualfund
    code: 119533
`
	log.Info("Generating config file: ", configFilePath)
	journalFilePath := filepath.Join(cwd, "personal.ledger")
	dbFilePath := filepath.Join(cwd, "paisa.db")
	err := ioutil.WriteFile(configFilePath, []byte(fmt.Sprintf(config, journalFilePath, dbFilePath)), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func loadPrices(schemeCode string, commodityName string, pricesTree map[string]*btree.BTree) {
	prices, err := mutualfund.GetNav(schemeCode, commodityName)
	if err != nil {
		log.Fatal(err)
	}

	pricesTree[commodityName] = btree.New(2)
	for _, price := range prices {
		pricesTree[commodityName].ReplaceOrInsert(*price)
	}
}

func formatFloat(num float64) string {
	s := fmt.Sprintf("%.4f", num)
	return strings.TrimRight(strings.TrimRight(s, "0"), ".")
}

func emitSalary(file *os.File, start time.Time) {
	var salary float64 = 100000 + (float64(start.Year())-START_YEAR)*(100000*0.05)
	var company string
	if start.Year() > 2017 {
		company = "Globex"
	} else {
		company = "Acme"
	}

	_, err := file.WriteString(fmt.Sprintf(`
%s Salary
    Income:Salary:%s
    Assets:Debt:EPF                 %s INR
    Tax                             %s INR
    Checking                        %s INR
`, start.Format("2006/01/02"), company, formatFloat(salary*0.12), formatFloat(salary*0.20), formatFloat(salary*0.68)))
	if err != nil {
		log.Fatal(err)
	}

	if start.Year() > START_YEAR && start.Month() == time.March {
		_, err = file.WriteString(fmt.Sprintf(`
%s EPF Interest
    Income:Interest:EPF
    Assets:Debt:EPF                 %s INR
`, start.Format("2006/01/02"), formatFloat(salary*0.12*((float64(start.Year())-START_YEAR)*12)*0.075)))

		if err != nil {
			log.Fatal(err)
		}

	}

}

func emitEquityMutualFund(file *os.File, start time.Time, pricesTree map[string]*btree.BTree) {
	multiplier := 1.0
	if start.Year() > START_YEAR+1 && rand.Intn(3) == 0 {
		multiplier = -1.0
	}
	pc := utils.BTreeDescendFirstLessOrEqual(pricesTree["NIFTY"], price.Price{Date: start})
	_, err := file.WriteString(fmt.Sprintf(`
%s Mutual Fund Nifty
    Assets:Equity:NIFTY  %s NIFTY @ %s INR
    Checking
`, start.Format("2006/01/02"), formatFloat(10000/pc.Value*multiplier), formatFloat(pc.Value)))
	if err != nil {
		log.Fatal(err)
	}

	pc = utils.BTreeDescendFirstLessOrEqual(pricesTree["NIFTY_JR"], price.Price{Date: start})
	_, err = file.WriteString(fmt.Sprintf(`
%s Mutual Fund Nifty Next 50
    Assets:Equity:NIFTY_JR  %s NIFTY_JR @ %s INR
    Checking
`, start.Format("2006/01/02"), formatFloat(10000/pc.Value*multiplier), formatFloat(pc.Value)))
	if err != nil {
		log.Fatal(err)
	}

}

func emitDebtMutualFund(file *os.File, start time.Time, pricesTree map[string]*btree.BTree) {
	multiplier := 1.0
	if start.Year() > START_YEAR+1 && rand.Intn(3) == 0 {
		multiplier = -1.0
	}
	pc := utils.BTreeDescendFirstLessOrEqual(pricesTree["ABCBF"], price.Price{Date: start})
	_, err := file.WriteString(fmt.Sprintf(`
%s Mutual Fund Birla Corporate Fund
    Assets:Debt:ABCBF  %s ABCBF @ %s INR
    Checking
`, start.Format("2006/01/02"), formatFloat(10000/pc.Value*multiplier), formatFloat(pc.Value)))
	if err != nil {
		log.Fatal(err)
	}
}

func generateJournalFile(cwd string) {
	journalFilePath := filepath.Join(cwd, "personal.ledger")
	log.Info("Generating journal file: ", journalFilePath)
	ledgerFile, err := os.OpenFile(journalFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}

	end := time.Now()
	start, err := time.Parse("02-01-2006", fmt.Sprintf("01-01-%d", START_YEAR))
	if err != nil {
		log.Fatal(err)
	}

	pricesTree := make(map[string]*btree.BTree)
	loadPrices("120716", "NIFTY", pricesTree)
	loadPrices("120684", "NIFTY_JR", pricesTree)
	loadPrices("119533", "ABCBF", pricesTree)

	for ; start.Before(end); start = start.AddDate(0, 1, 0) {
		emitSalary(ledgerFile, start)
		emitEquityMutualFund(ledgerFile, start, pricesTree)
		emitDebtMutualFund(ledgerFile, start, pricesTree)
	}
}
