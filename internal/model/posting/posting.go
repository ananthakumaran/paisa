package posting

import (
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Posting struct {
	ID                   uint      `gorm:"primaryKey" json:"id"`
	TransactionID        string    `json:"transaction_id"`
	Date                 time.Time `json:"date"`
	Payee                string    `json:"payee"`
	Account              string    `json:"account"`
	Commodity            string    `json:"commodity"`
	Quantity             float64   `json:"quantity"`
	Amount               float64   `json:"amount"`
	Status               string    `json:"status"`
	TagRecurring         string    `json:"tag_recurring"`
	TransactionBeginLine uint64    `json:"transaction_begin_line"`
	TransactionEndLine   uint64    `json:"transaction_end_line"`
	FileName             string    `json:"file_name"`

	MarketAmount float64 `gorm:"-:all" json:"market_amount"`
}

func (p Posting) GroupDate() time.Time {
	return p.Date
}

func (p *Posting) RestName(level int) string {
	return strings.Join(strings.Split(p.Account, ":")[level:], ":")
}

func (p Posting) Negate() Posting {
	clone := p
	clone.Quantity = -p.Quantity
	clone.Amount = -p.Amount
	return clone
}

func (p *Posting) Price() float64 {
	return p.Amount / p.Quantity
}

func (p *Posting) AddAmount(amount float64) {
	p.Amount += amount
}

func (p *Posting) AddQuantity(quantity float64) {
	price := p.Price()
	p.Quantity += quantity
	p.Amount = p.Quantity * price
}

func (p Posting) WithQuantity(quantity float64) Posting {
	clone := p
	clone.Quantity = quantity
	clone.Amount = quantity * p.Price()
	return clone
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
