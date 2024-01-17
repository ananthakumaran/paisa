package generator

import (
	"fmt"
	"math"
	"os"
	"path/filepath"

	"time"

	"strings"

	"math/rand"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/scraper/mutualfund"
	"github.com/ananthakumaran/paisa/internal/scraper/nps"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/google/btree"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
)

const START_YEAR = 2014

type GeneratorState struct {
	Balance      float64
	EPFBalance   float64
	Ledger       *os.File
	YearlySalary float64
	Rent         float64
	LoanBalance  float64
	NiftyBalance float64
}

var pricesTree map[string]*btree.BTree

func MinimalConfig(cwd string) {
	configFilePath := filepath.Join(cwd, "paisa.yaml")
	config := `
journal_path: '%s'
db_path: '%s'
`
	log.Info("Generating config file: ", configFilePath)
	journalFilePath := filepath.Join(cwd, "main.ledger")
	dbFilePath := filepath.Join(cwd, "paisa.db")
	err := os.WriteFile(configFilePath, []byte(fmt.Sprintf(config, filepath.Base(journalFilePath), filepath.Base(dbFilePath))), 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Generating journal file: ", journalFilePath)
	_, err = os.OpenFile(journalFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func Demo(cwd string) {
	generateConfigFile(cwd)
	generateJournalFile(cwd)
	generateSheetFile(cwd)
}

func generateConfigFile(cwd string) {
	configFilePath := filepath.Join(cwd, "paisa.yaml")
	config := `
journal_path: '%s'
db_path: '%s'
ledger_cli: ledger
default_currency: INR
goals:
  retirement:
    - name: Early Retirement
      icon: mdi:palm-tree
      swr: 3
      savings:
        - Assets:Debt:*
        - Assets:Equity:*
      expenses:
        - Expenses:Rent
        - Expenses:Utilities
        - Expenses:Shopping
        - Expenses:Restaurants
        - Expenses:Food
        - Expenses:Interest:*
  savings:
    - name: Millionaire
      icon: mdi:car-sports
      target: 80000000
      target_date: "2036-01-01"
      rate: 10
      accounts:
        - '!Assets:Checking:SBI'
allocation_targets:
  - name: Debt
    target: 40
    accounts:
      - Assets:Debt:*
  - name: Equity
    target: 60
    accounts:
      - Assets:Equity:*
schedule_al:
  - code: bank
    accounts:
      - Assets:Checking:SBI
  - code: share
    accounts:
      - Assets:Equity:*
      - Assets:Debt:*
  - code: liability
    accounts:
      - Liabilities:Homeloan
  - code: immovable
    accounts:
      - Assets:House
commodities:
  - name: NIFTY
    type: mutualfund
    price:
      provider: in-mfapi
      code: 120716
    harvest: 365
    tax_category: equity65
  - name: PPFAS
    type: mutualfund
    price:
      provider: in-mfapi
      code: 122639
    harvest: 365
    tax_category: equity65
  - name: ABCBF
    type: mutualfund
    price:
      provider: in-mfapi
      code: 119533
    harvest: 1095
    tax_category: debt
  - name: NPS_HDFC_E
    type: nps
    price:
      provider: com-purifiedbytes-nps
      code: SM008001
  - name: NPS_HDFC_C
    type: nps
    price:
      provider: com-purifiedbytes-nps
      code: SM008002
  - name: NPS_HDFC_G
    type: nps
    price:
      provider: com-purifiedbytes-nps
      code: SM008003
`
	log.Info("Generating config file: ", configFilePath)
	journalFilePath := filepath.Join(cwd, "main.ledger")
	dbFilePath := filepath.Join(cwd, "paisa.db")
	err := os.WriteFile(configFilePath, []byte(fmt.Sprintf(config, filepath.Base(journalFilePath), filepath.Base(dbFilePath))), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func emitTransaction(file *os.File, date time.Time, payee string, from string, to string, amount interface{}) {
	amountString := ""
	switch amount.(type) {
	case string:
		amountString = amount.(string)
	case float64:
		amountString = formatFloat(amount.(float64))
	}

	_, err := file.WriteString(fmt.Sprintf(`
%s %s
    %s                                %s INR
    %s
`, date.Format("2006/01/02"), payee, to, amountString, from))
	if err != nil {
		log.Fatal(err)
	}
}

func emitCommodityBuy(file *os.File, date time.Time, commodity string, from string, to string, amount float64) float64 {
	pc := utils.BTreeDescendFirstLessOrEqual(pricesTree[commodity], price.Price{Date: date})
	units := amount / pc.Value.InexactFloat64()
	_, err := file.WriteString(fmt.Sprintf(`
%s Investment
    %s                      %s %s @    %s INR
    %s
`, date.Format("2006/01/02"), to, formatFloat(units), commodity, formatFloat(pc.Value.InexactFloat64()), from))
	if err != nil {
		log.Fatal(err)
	}
	return units
}

func emitCommoditySell(file *os.File, date time.Time, commodity string, from string, to string, amount float64, availableUnits float64) (float64, float64) {
	pc := utils.BTreeDescendFirstLessOrEqual(pricesTree[commodity], price.Price{Date: date})
	requiredUnits := amount / pc.Value.InexactFloat64()
	units := math.Min(availableUnits, requiredUnits)
	return emitCommodityBuy(file, date, commodity, from, to, -units*pc.Value.InexactFloat64()), units * pc.Value.InexactFloat64()
}

func loadPrices(schemeCode string, commodityType config.CommodityType, commodityName string, pricesTree map[string]*btree.BTree) {
	var prices []*price.Price
	var err error

	switch commodityType {
	case config.MutualFund:
		prices, err = mutualfund.GetNav(schemeCode, commodityName)
	case config.NPS:
		prices, err = nps.GetNav(schemeCode, commodityName)
	}

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

func roundToK(amount float64) float64 {
	if amount < 20000 {
		return float64(int(amount/100) * 100)
	}
	return float64(int(amount/1000) * 1000)
}

func incrementByPercentRange(amount float64, min int, max int) float64 {
	return roundToK(amount + amount*percentRange(min, max))
}

func percentRange(min int, max int) float64 {
	if min == max {
		return float64(min) * 0.01
	}
	return float64(randRange(min, max)) * 0.01
}

func randRange(min int, max int) int {
	return rand.Intn(max-min) + min
}

func taxRate(amount float64) float64 {
	if amount < 500000 {
		return 0
	} else if amount < 750000 {
		return 0.10
	} else if amount < 1000000 {
		return 0.15
	} else if amount < 1250000 {
		return 0.20
	} else if amount < 1500000 {
		return 0.25
	}
	return 0.30
}

func emitChitFund(state *GeneratorState) {
	start, _ := time.Parse("02-01-2006", "01-01-2016")
	end, _ := time.Parse("02-01-2006", "01-11-2016")

	for ; start.Before(end); start = start.AddDate(0, 1, 0) {
		price := 10000 - ((time.November - start.Month()) * 100)
		amount := fmt.Sprintf("1 CHIT @ %d", price)

		if start.Month() >= time.June {
			emitTransaction(state.Ledger, start, "Chit installment", "Assets:Checking:SBI", "Liabilities:Chit", amount)
		} else {
			emitTransaction(state.Ledger, start, "Chit installment", "Assets:Checking:SBI", "Assets:Debt:Chit", amount)
		}

		if start.Month() == time.June {
			amount = fmt.Sprintf("-5 CHIT @ %d", price)
			emitTransaction(state.Ledger, start, "Chit withdraw", "Assets:Checking:SBI", "Assets:Debt:Chit", amount)
			amount = fmt.Sprintf("-5 CHIT @ %d", price)
			emitTransaction(state.Ledger, start, "Chit withdraw", "Assets:Checking:SBI", "Liabilities:Chit", amount)
		}

	}
}

func emitSalary(state *GeneratorState, start time.Time) {
	if start.Month() == time.April {
		state.YearlySalary = incrementByPercentRange(state.YearlySalary, 10, 15)
	}

	var salary float64 = state.YearlySalary / 12
	var company string
	if start.Year() > 2017 {
		company = "Globex"
	} else {
		company = "Acme"
	}

	tax := salary * taxRate(state.YearlySalary)
	epf := salary * 0.12
	nps := salary * 0.10
	state.EPFBalance += epf
	netSalary := salary - tax - epf - nps
	state.Balance += netSalary

	salaryAccount := fmt.Sprintf("Income:Salary:%s", company)
	emitTransaction(state.Ledger, start, "Salary", salaryAccount, "Assets:Checking:SBI", netSalary)
	emitTransaction(state.Ledger, start, "Salary EPF", salaryAccount, "Assets:Debt:EPF", epf)
	emitTransaction(state.Ledger, start, "Salary Tax", salaryAccount, "Expenses:Tax", tax)
	emitCommodityBuy(state.Ledger, start, "NPS_HDFC_E", salaryAccount, "Assets:Debt:NPS:HDFC:E", nps*0.75)
	emitCommodityBuy(state.Ledger, start, "NPS_HDFC_C", salaryAccount, "Assets:Equity:NPS:HDFC:C", nps*0.15)
	emitCommodityBuy(state.Ledger, start, "NPS_HDFC_G", salaryAccount, "Assets:Equity:NPS:HDFC:G", nps*0.10)

}

func emitExpense(state *GeneratorState, start time.Time) {
	if start.Month() == time.April {
		state.Rent = incrementByPercentRange(state.Rent, 5, 10)
	}

	emit := func(payee string, account string, amount float64, fuzz float64) {
		actualAmount := roundToK(percentRange(int(fuzz*100), 100) * amount)
		start = start.AddDate(0, 0, 1)
		emitTransaction(state.Ledger, start, payee, "Assets:Checking:SBI", account, actualAmount)
		state.Balance -= actualAmount
	}

	emit("Rent", "Expenses:Rent", state.Rent, 1.0)
	emit("Internet", "Expenses:Utilities", 1500, 1.0)
	emit("Mobile", "Expenses:Utilities", 430, 1.0)
	emit("Shopping", "Expenses:Shopping", 3000, 0.5)
	emit("Eat out", "Expenses:Restaurants", 2500, 0.5)
	emit("Groceries", "Expenses:Food", 5000, 0.9)

	if state.LoanBalance > 0 {
		emi := math.Min(state.Balance-10000, 30000.0)
		interest := (state.LoanBalance * 0.08 / 12)
		principal := emi - interest
		state.LoanBalance -= principal
		emit("EMI", "Expenses:Interest:Homeloan", interest, 1.0)
		emit("EMI", "Liabilities:Homeloan", principal, 1.0)
	}

	if state.Balance < 10000 {
		return
	}

	if lo.Contains([]time.Month{time.January, time.April, time.November, time.December}, start.Month()) {
		emit("Dress", "Expenses:Clothing", 5000, 0.5)
	}
}
func emitInvestment(state *GeneratorState, start time.Time) {
	if start.Month() == time.April {
		epfInterest := state.EPFBalance * 0.08
		emitTransaction(state.Ledger, start, "EPF Interest", "Income:Interest:EPF", "Assets:Debt:EPF", epfInterest)
		state.EPFBalance += epfInterest
	}

	if state.Balance < 10000 {
		return
	}

	equity1 := roundToK(state.Balance * 0.5)
	equity2 := roundToK(state.Balance * 0.2)
	debt := roundToK(state.Balance * 0.3)

	state.Balance -= equity1
	state.NiftyBalance += emitCommodityBuy(state.Ledger, start, "NIFTY", "Assets:Checking:SBI", "Assets:Equity:NIFTY", equity1)

	state.Balance -= equity2
	emitCommodityBuy(state.Ledger, start, "PPFAS", "Assets:Checking:SBI", "Assets:Equity:PPFAS", equity2)

	state.Balance -= debt
	emitCommodityBuy(state.Ledger, start, "ABCBF", "Assets:Checking:SBI", "Assets:Debt:ABCBF", debt)

	if start.Month() == time.March {
		units, amount := emitCommoditySell(state.Ledger, start.AddDate(0, 0, 15), "NIFTY", "Assets:Checking:SBI", "Assets:Equity:NIFTY", 75000, state.NiftyBalance)
		state.NiftyBalance += units
		state.Balance += amount
	}

}

func generateJournalFile(cwd string) {
	journalFilePath := filepath.Join(cwd, "main.ledger")
	log.Info("Generating journal file: ", journalFilePath)
	ledgerFile, err := os.OpenFile(journalFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}

	startMonth := utils.BeginningOfMonth(utils.EndOfToday())
	endMonth := startMonth.AddDate(0, 2, 0)

	_, err = ledgerFile.WriteString(`
= Expenses:Rent
    ; Recurring: Rent
    ; Period: 1 * ?

= expr payee=~/Internet/
    ; Recurring: Internet
    ; Period: 1 * ?

= expr payee=~/EPF/
    ; Recurring: EPF

= Liabilities:Homeloan
    ; Recurring: EMI Principle

= Expenses:Interest:Homeloan
    ; Recurring: EMI Interest

~ Monthly from `)

	_, err = ledgerFile.WriteString(fmt.Sprintf("%s to %s", startMonth.Format("2006-01-02"), endMonth.Format("2006-01-02")))

	_, err = ledgerFile.WriteString(`
    Expenses:Rent                              15000 INR
    Expenses:Interest:Homeloan                  6000 INR
    Expenses:Food                               5000 INR
    Expenses:Utilities                          2000 INR
    Expenses:Shopping                           3000 INR
    Expenses:Clothing                           1000 INR
    Assets:Checking:SBI

`)
	if err != nil {
		log.Fatal(err)
	}

	end := utils.EndOfToday()
	start, err := time.Parse("02-01-2006", fmt.Sprintf("01-01-%d", START_YEAR))
	if err != nil {
		log.Fatal(err)
	}

	pricesTree = make(map[string]*btree.BTree)
	loadPrices("120716", config.MutualFund, "NIFTY", pricesTree)
	loadPrices("122639", config.MutualFund, "PPFAS", pricesTree)
	loadPrices("119533", config.MutualFund, "ABCBF", pricesTree)
	loadPrices("SM008001", config.NPS, "NPS_HDFC_E", pricesTree)
	loadPrices("SM008002", config.NPS, "NPS_HDFC_C", pricesTree)
	loadPrices("SM008003", config.NPS, "NPS_HDFC_G", pricesTree)

	state := GeneratorState{Balance: 0, Ledger: ledgerFile, YearlySalary: 1000000, Rent: 10000, LoanBalance: 2500000}

	emitTransaction(state.Ledger, start, "Home purchase", "Liabilities:Homeloan", "Assets:House", "1 APT @ 2500000")

	for ; start.Before(end); start = start.AddDate(0, 1, 0) {
		emitSalary(&state, start)
		emitExpense(&state, start)
		emitInvestment(&state, start)
	}

	emitChitFund(&state)
}

func generateSheetFile(cwd string) {
	sheetFilePath := filepath.Join(cwd, "Schedule AL.paisa")
	sheet := `
date_query = {date <= [2023-03-31]}
cost_basis(x) = cost(fifo(x AND date_query))
cost_basis_negative(x) = cost(fifo(negate(x AND date_query)))

# Immovable
immovable = cost_basis({account = Assets:House})

# Movable
metal = 0
art = 0
vehicle = 0
bank = cost_basis({account =~ /^Assets:Checking:SBI/})
share = cost_basis({account =~ /^Assets:Equity:.*/ OR
                    account =~ /^Assets:Debt:.*/})
insurance = 0
loan = 0
cash = 0

# Liability
liability = cost_basis_negative({account =~ /^Liabilities:Homeloan/})

# Total
total = immovable + metal + art + vehicle + bank + share + insurance + loan + cash - liability
`
	log.Info("Generating sheet file: ", sheetFilePath)
	err := os.WriteFile(sheetFilePath, []byte(sheet), 0644)
	if err != nil {
		log.Fatal(err)
	}
}
