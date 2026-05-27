package server

import (
	"strings"
	"time"

	"github.com/ananthakumaran/paisa/internal/accounting"
	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/account"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type InvestmentYearlyCard struct {
	StartDate         time.Time         `json:"start_date"`
	EndDate           time.Time         `json:"end_date"`
	Postings          []posting.Posting `json:"postings"`
	GrossSalaryIncome decimal.Decimal   `json:"gross_salary_income"`
	GrossOtherIncome  decimal.Decimal   `json:"gross_other_income"`
	NetTax            decimal.Decimal   `json:"net_tax"`
	NetIncome         decimal.Decimal   `json:"net_income"`
	NetInvestment     decimal.Decimal   `json:"net_investment"`
	NetExpense        decimal.Decimal   `json:"net_expense"`
	SavingsRate       decimal.Decimal   `json:"savings_rate"`
}

// filterOutInternalTransfers removes postings that belong to a
// transaction classified as an internal transfer under the user's
// `transfer_accounts` config. To classify correctly we need the FULL
// transaction (including legs that may not be present in the filtered
// `assets` slice — e.g. an Assets:Checking leg that was excluded
// upstream), so we fetch every posting that shares a transaction id
// with one in `assets`.
func filterOutInternalTransfers(db *gorm.DB, assets []posting.Posting) []posting.Posting {
	transferGlobs := config.GetConfig().TransferAccounts
	if len(transferGlobs) == 0 || len(assets) == 0 {
		return assets
	}

	txIDs := lo.Uniq(lo.Map(assets, func(p posting.Posting, _ int) string { return p.TransactionID }))
	allLegs := query.Init(db).Where("transaction_id in ?", txIDs).All()
	byTx := lo.GroupBy(allLegs, func(p posting.Posting) string { return p.TransactionID })

	return lo.Filter(assets, func(p posting.Posting, _ int) bool {
		ps, ok := byTx[p.TransactionID]
		if !ok {
			return true
		}
		return !accounting.IsInternalTransfer(ps, transferGlobs)
	})
}

func GetInvestment(db *gorm.DB) gin.H {
	assets := query.Init(db).Like("Assets:%").NotAccountPrefix("Assets:Checking").
		Where("transaction_id not in (select transaction_id from postings p where p.account like ? and p.transaction_id = transaction_id)", "Liabilities:%").
		All()
	incomes := query.Init(db).Like("Income:%").All()
	expenses := query.Init(db).Like("Expenses:%").All()
	p := query.Init(db).First()

	if p == nil {
		return gin.H{"assets": []posting.Posting{}, "yearly_cards": []InvestmentYearlyCard{}}
	}

	assets = lo.Filter(assets, func(p posting.Posting, _ int) bool { return !service.IsStockSplit(db, p) })
	assets = filterOutInternalTransfers(db, assets)
	assets = filterOutNonInvestmentKinds(assets, toAccountLookup(config.GetConfig().Accounts))
	return gin.H{"assets": assets, "yearly_cards": computeInvestmentYearlyCard(p.Date, assets, expenses, incomes)}
}

// isInvestmentKind reports whether an AccountKind represents a flow that
// counts as a new investment for the Investment Timeline and Savings Rate
// computations.
//
// Investment kinds:
//
//	mutual_fund, stock, bond, structured_deposit,
//	tax_deferred_fund, crypto
//
// Everything else — including cash-like accounts (bank_current,
// bank_savings, cash_equivalent), durable assets (real_estate, vehicle),
// the cash portion of PPA (tax_deferred_cash), and receivables / housing
// fund — is excluded so that mere reshuffling between cash accounts is
// not counted as "investment".
func isInvestmentKind(k account.AccountKind) bool {
	switch k {
	case account.MutualFund,
		account.Stock,
		account.Bond,
		account.StructuredDeposit,
		account.TaxDeferredFund,
		account.Crypto:
		return true
	}
	return false
}

// filterOutNonInvestmentKinds keeps only postings whose account resolves
// to an investment kind (see isInvestmentKind). Pass `accounts` to honor
// user-configured `accounts[].kind` overrides; without it the function
// falls back to M1-D's path-prefix rules in account.GetKind.
func filterOutNonInvestmentKinds(postings []posting.Posting, accounts []account.Account) []posting.Posting {
	if len(postings) == 0 {
		return postings
	}
	return lo.Filter(postings, func(p posting.Posting, _ int) bool {
		return isInvestmentKind(account.GetKind(p.Account, accounts))
	})
}

func computeInvestmentYearlyCard(start time.Time, assets []posting.Posting, expenses []posting.Posting, incomes []posting.Posting) []InvestmentYearlyCard {
	var yearlyCards []InvestmentYearlyCard = make([]InvestmentYearlyCard, 0)

	if len(assets) == 0 {
		return yearlyCards
	}

	var p posting.Posting
	end := utils.EndOfToday()
	for start = utils.BeginningOfCalendarYear(start); start.Before(end); start = start.AddDate(1, 0, 0) {
		yearEnd := utils.EndOfCalendarYear(start)
		var currentYearPostings []posting.Posting = make([]posting.Posting, 0)
		for len(assets) > 0 && utils.IsWithDate(assets[0].Date, start, yearEnd) {
			p, assets = assets[0], assets[1:]
			currentYearPostings = append(currentYearPostings, p)
		}

		var currentYearTaxes []posting.Posting = make([]posting.Posting, 0)
		var currentYearExpenses []posting.Posting = make([]posting.Posting, 0)

		for len(expenses) > 0 && utils.IsWithDate(expenses[0].Date, start, yearEnd) {
			p, expenses = expenses[0], expenses[1:]
			if utils.IsSameOrParent(p.Account, "Expenses:Tax") {
				currentYearTaxes = append(currentYearTaxes, p)
			} else {
				currentYearExpenses = append(currentYearExpenses, p)
			}
		}

		netTax := accounting.CostSum(currentYearTaxes)
		netExpense := accounting.CostSum(currentYearExpenses)

		var currentYearIncomes []posting.Posting = make([]posting.Posting, 0)
		for len(incomes) > 0 && utils.IsWithDate(incomes[0].Date, start, yearEnd) {
			p, incomes = incomes[0], incomes[1:]
			currentYearIncomes = append(currentYearIncomes, p)
		}

		grossSalaryIncome := utils.SumBy(currentYearIncomes, func(p posting.Posting) decimal.Decimal {
			if strings.HasPrefix(p.Account, "Income:Salary") {
				return p.Amount.Neg()
			} else {
				return decimal.Zero
			}
		})
		grossOtherIncome := utils.SumBy(currentYearIncomes, func(p posting.Posting) decimal.Decimal {
			if !strings.HasPrefix(p.Account, "Income:Salary") {
				return p.Amount.Neg()
			} else {
				return decimal.Zero
			}
		})

		netInvestment := accounting.CostSum(currentYearPostings)

		netIncome := grossSalaryIncome.Add(grossOtherIncome).Sub(netTax)
		var savingsRate decimal.Decimal = decimal.Zero
		if !netIncome.Equal(decimal.Zero) {
			savingsRate = (netInvestment.Div(netIncome)).Mul(decimal.NewFromInt(100))
		}

		yearlyCards = append(yearlyCards, InvestmentYearlyCard{
			StartDate:         start,
			EndDate:           yearEnd,
			Postings:          currentYearPostings,
			NetTax:            netTax,
			GrossSalaryIncome: grossSalaryIncome,
			GrossOtherIncome:  grossOtherIncome,
			NetIncome:         netIncome,
			NetInvestment:     netInvestment,
			NetExpense:        netExpense,
			SavingsRate:       savingsRate,
		})

	}
	return yearlyCards
}
