package posting

import (
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

type Posting struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Date      time.Time `json:"date"`
	Payee     string    `json:"payee"`
	Account   string    `json:"account"`
	Commodity string    `json:"commodity"`
	Quantity  float64   `json:"quantity"`
	Amount    float64   `json:"amount"`

	MarketAmount float64 `gorm:"-:all" json:"market_amount"`
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
