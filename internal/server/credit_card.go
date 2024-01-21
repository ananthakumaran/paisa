package server

import (
	"time"

	"github.com/ananthakumaran/paisa/internal/accounting"
	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/transaction"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
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
	StatementStartDate   time.Time       `json:"statementStartDate"`
	StatementEndDate     time.Time       `json:"statementEndDate"`
	DueDate              time.Time       `json:"dueDate"`
	PaidDate             *time.Time      `json:"paidDate"`
	Credits              decimal.Decimal `json:"credits"`
	Debits               decimal.Decimal `json:"debits"`
	DebitsRunningBalance decimal.Decimal
	OpeningBalance       decimal.Decimal           `json:"openingBalance"`
	ClosingBalance       decimal.Decimal           `json:"closingBalance"`
	Postings             []posting.Posting         `json:"postings"`
	Transactions         []transaction.Transaction `json:"transactions"`
}

func GetCreditCards(db *gorm.DB) gin.H {
	creditCards := []CreditCardSummary{}

	for _, creditCardConfig := range config.GetConfig().CreditCards {
		ps := query.Init(db).Where("account = ?", creditCardConfig.Account).All()
		creditCards = append(creditCards, buildCreditCard(db, creditCardConfig, ps, false))
	}

	return gin.H{"creditCards": creditCards}
}

func GetCreditCard(db *gorm.DB, account string) gin.H {
	for _, creditCardConfig := range config.GetConfig().CreditCards {
		if creditCardConfig.Account == account {
			ps := query.Init(db).Where("account = ?", creditCardConfig.Account).All()
			creditCard := buildCreditCard(db, creditCardConfig, ps, true)
			return gin.H{"creditCard": creditCard, "found": true}
		}
	}

	return gin.H{"found": false}
}

func buildCreditCard(db *gorm.DB, creditCardConfig config.CreditCard, ps []posting.Posting, includePostings bool) CreditCardSummary {
	bills := computeBills(db, creditCardConfig, ps, includePostings)
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

func computeBills(db *gorm.DB, creditCardConfig config.CreditCard, ps []posting.Posting, includePostings bool) []CreditCardBill {
	bills := []CreditCardBill{}

	grouped := accounting.GroupByMonthlyBillingCycle(ps, creditCardConfig.StatementEndDay)

	balance := decimal.Zero
	creditsRunningBalance := decimal.Zero
	debitsRunningBalance := decimal.Zero
	unpaidBill := 0

	for _, month := range utils.SortedKeys(grouped) {
		statementEndDate, err := time.Parse("2006-01", month)
		if err != nil {
			log.Fatal(err)
		}

		statementEndDate = statementEndDate.AddDate(0, 0, creditCardConfig.StatementEndDay-1)
		statementStartDate := statementEndDate.AddDate(0, -1, 1)

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
			Transactions:       []transaction.Transaction{},
		}

		transactionIDs := map[string]bool{}

		for _, p := range grouped[month] {
			balance = balance.Add(p.Amount.Neg())

			if p.Amount.IsPositive() {
				creditsRunningBalance = creditsRunningBalance.Add(p.Amount)
				bill.Credits = bill.Credits.Add(p.Amount)
				for unpaidBill < len(bills) {
					if bills[unpaidBill].DebitsRunningBalance.LessThanOrEqual(creditsRunningBalance) {
						paidDate := p.Date
						bills[unpaidBill].PaidDate = &paidDate
						unpaidBill++
					} else {
						break
					}
				}
			} else {
				bill.Debits = bill.Debits.Add(p.Amount.Neg())
				debitsRunningBalance = debitsRunningBalance.Add(p.Amount.Neg())
			}

			if includePostings {
				bill.Postings = append(bill.Postings, p)
				transactionIDs[p.TransactionID] = true
			}

		}

		bill.DebitsRunningBalance = debitsRunningBalance
		bill.ClosingBalance = balance
		bill.Transactions = lo.Map(lo.Keys(transactionIDs), func(id string, _ int) transaction.Transaction {
			t, _ := transaction.GetById(db, id)
			return t
		})
		accounting.SortTransactionAsc(bill.Transactions)
		bills = append(bills, bill)
	}

	return bills
}
