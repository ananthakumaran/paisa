package server

import (
	"time"

	"github.com/ananthakumaran/paisa/internal/accounting"
	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type CreditCardSummary struct {
	Account     string           `json:"account"`
	Network     string           `json:"network"`
	Number      string           `json:"number"`
	Balance     decimal.Decimal  `json:"balance"`
	Bills       []CreditCardBill `json:"bills"`
	CreditLimit decimal.Decimal  `json:"creditLimit"`
}

type CreditCardBill struct {
	StatementStartDate time.Time         `json:"statementStartDate"`
	StatementEndDate   time.Time         `json:"statementEndDate"`
	DueDate            time.Time         `json:"dueDate"`
	PaidDate           *time.Time        `json:"paidDate"`
	Credits            decimal.Decimal   `json:"credits"`
	Debits             decimal.Decimal   `json:"debits"`
	OpeningBalance     decimal.Decimal   `json:"openingBalance"`
	ClosingBalance     decimal.Decimal   `json:"closingBalance"`
	Postings           []posting.Posting `json:"postings"`
}

func GetCreditCards(db *gorm.DB) gin.H {
	creditCards := []CreditCardSummary{}

	for _, creditCardConfig := range config.GetConfig().CreditCards {
		ps := query.Init(db).Where("account = ?", creditCardConfig.Account).All()
		creditCards = append(creditCards, buildCreditCard(creditCardConfig, ps))
	}

	return gin.H{"creditCards": creditCards}
}

func buildCreditCard(creditCardConfig config.CreditCard, ps []posting.Posting) CreditCardSummary {
	bills := computeBills(creditCardConfig, ps)
	balance := decimal.Zero
	if len(bills) > 0 {
		balance = bills[len(bills)-1].ClosingBalance
	}
	return CreditCardSummary{
		Account:     creditCardConfig.Account,
		Network:     creditCardConfig.Network,
		Number:      creditCardConfig.Number,
		Balance:     balance,
		Bills:       bills,
		CreditLimit: decimal.NewFromInt(int64(creditCardConfig.CreditLimit)),
	}
}

func computeBills(creditCardConfig config.CreditCard, ps []posting.Posting) []CreditCardBill {
	bills := []CreditCardBill{}

	grouped := accounting.GroupByMonthlyBillingCycle(ps, creditCardConfig.StatementEndDay)

	balance := decimal.Zero
	unpaidBalance := decimal.Zero
	unpaidBill := 0

	for _, month := range utils.SortedKeys(grouped) {
		statementEndDate, err := time.Parse("2006-01", month)
		if err != nil {
			log.Fatal(err)
		}

		statementEndDate = statementEndDate.AddDate(0, 0, creditCardConfig.StatementEndDay-1)
		statementStartDate := statementEndDate.AddDate(0, -1, -1)

		var dueDate time.Time
		if creditCardConfig.StatementEndDay < creditCardConfig.DueDay {
			dueDate = utils.BeginningOfMonth(statementEndDate).AddDate(0, 0, creditCardConfig.DueDay-1)
		} else {
			dueDate = utils.BeginningOfMonth(statementEndDate).AddDate(0, 1, creditCardConfig.DueDay-1)
		}

		bill := CreditCardBill{
			StatementStartDate: statementStartDate,
			StatementEndDate:   statementEndDate,
			DueDate:            dueDate,
			OpeningBalance:     balance,
			Postings:           []posting.Posting{},
		}

		for _, p := range grouped[month] {
			balance = balance.Add(p.Amount.Neg())

			if p.Amount.IsPositive() {
				bill.Credits = bill.Credits.Add(p.Amount)
				if unpaidBalance.IsPositive() {

					unpaidBalance = unpaidBalance.Sub(p.Amount)
					if unpaidBalance.LessThanOrEqual(decimal.Zero) {

						unpaidBalance = decimal.Zero
						for i := unpaidBill; i < len(bills); i++ {
							paidDate := p.Date
							bills[i].PaidDate = &paidDate
						}
						unpaidBill = len(bills)
					}

				}
			} else {
				bill.Debits = bill.Debits.Add(p.Amount.Neg())
			}

			bill.Postings = append(bill.Postings, p)

		}

		bill.ClosingBalance = balance
		unpaidBalance = balance
		bills = append(bills, bill)
	}

	return bills
}
