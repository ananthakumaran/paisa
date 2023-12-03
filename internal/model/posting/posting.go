package posting

import (
	"strings"
	"time"

	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	ASSETS               = "assets"
	ASSETS_CASH          = "assets:cash"
	INCOME               = "income"
	INCOME_INTEREST      = "income:interest"
	INCOME_DIVIDEND      = "income:dividend"
	INCOME_CAPITAL_GAINS = "income:capital_gains"
	EXPENSES             = "expenses"
	EXPENSES_CHARGES     = "expenses:charges"
	EXPENSES_TAXES       = "expenses:taxes"
	LIABILITIES          = "liabilities"
)

type Posting struct {
	ID                   uint            `gorm:"primaryKey" json:"id"`
	TransactionID        string          `json:"transaction_id"`
	Date                 time.Time       `json:"date"`
	Payee                string          `json:"payee"`
	Account              string          `json:"account"`
	Commodity            string          `json:"commodity"`
	Quantity             decimal.Decimal `json:"quantity"`
	Amount               decimal.Decimal `json:"amount"`
	Status               string          `json:"status"`
	TagRecurring         string          `json:"tag_recurring"`
	TagPeriod            string          `json:"tag_period"`
	TransactionBeginLine uint64          `json:"transaction_begin_line"`
	TransactionEndLine   uint64          `json:"transaction_end_line"`
	FileName             string          `json:"file_name"`
	Forecast             bool            `json:"forecast"`
	Note                 string          `json:"note"`
	TransactionNote      string          `json:"transaction_note"`

	MarketAmount decimal.Decimal `gorm:"-:all" json:"market_amount"`
	Balance      decimal.Decimal `gorm:"-:all" json:"balance"`

	behaviours []string `gorm:"-:all"`
}

func (p Posting) GroupDate() time.Time {
	return p.Date
}

func (p *Posting) RestName(level int) string {
	return strings.Join(strings.Split(p.Account, ":")[level:], ":")
}

func (p Posting) Negate() Posting {
	clone := p
	clone.Quantity = p.Quantity.Neg()
	clone.Amount = p.Amount.Neg()
	return clone
}

func (p *Posting) Price() decimal.Decimal {
	if p.Quantity.IsZero() {
		return decimal.Zero
	}
	return p.Amount.Div(p.Quantity)
}

func (p *Posting) AddAmount(amount decimal.Decimal) {
	p.Amount = p.Amount.Add(amount)
}

func (p *Posting) AddQuantity(quantity decimal.Decimal) {
	price := p.Price()
	p.Quantity = p.Quantity.Add(quantity)
	p.Amount = p.Quantity.Mul(price)
}

func (p Posting) WithQuantity(quantity decimal.Decimal) Posting {
	clone := p
	clone.Quantity = quantity
	clone.Amount = quantity.Mul(p.Price())
	return clone
}

func (p Posting) Behaviours() []string {
	if p.behaviours == nil {
		p.behaviours = Behaviours(p.Account)
	}
	return p.behaviours
}

func (p Posting) HasBehaviour(behaviour string) bool {
	for _, b := range p.Behaviours() {
		if b == behaviour {
			return true
		}
	}
	return false
}

func UpsertAll(db *gorm.DB, postings []*Posting) {
	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Exec("DELETE FROM postings").Error
		if err != nil {
			return err
		}
		for _, posting := range postings {
			err := tx.Create(posting).Error
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}

func Behaviours(account string) []string {
	var behaviours []string
	if utils.IsParent(account, "Assets") {
		behaviours = append(behaviours, ASSETS)
	}

	if utils.IsSameOrParent(account, "Assets:Checking") {
		behaviours = append(behaviours, ASSETS_CASH)
	}

	if utils.IsParent(account, "Income") {
		behaviours = append(behaviours, INCOME)
	}

	if utils.IsSameOrParent(account, "Income:Interest") {
		behaviours = append(behaviours, INCOME_INTEREST)
	}

	if utils.IsSameOrParent(account, "Income:Dividend") {
		behaviours = append(behaviours, INCOME_DIVIDEND)
	}

	if utils.IsSameOrParent(account, "Income:Capital Gains") {
		behaviours = append(behaviours, INCOME_CAPITAL_GAINS)
	}

	if utils.IsParent(account, "Expenses") {
		behaviours = append(behaviours, EXPENSES)
	}

	if utils.IsSameOrParent(account, "Expenses:Charges") {
		behaviours = append(behaviours, EXPENSES_CHARGES)
	}

	if utils.IsSameOrParent(account, "Expenses:Tax") {
		behaviours = append(behaviours, EXPENSES_TAXES)
	}

	if utils.IsParent(account, "Liabilities") {
		behaviours = append(behaviours, LIABILITIES)
	}
	return behaviours
}
