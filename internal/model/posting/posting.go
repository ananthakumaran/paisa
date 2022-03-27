package posting

import (
	"gorm.io/gorm"
	"time"
)

type Posting struct {
	ID        uint `gorm:"primaryKey"`
	Date      time.Time
	Payee     string
	Account   string
	Commodity string
	Quantity  float64
}

func UpsertAll(db *gorm.DB, postings []*Posting) {
	db.Transaction(func(tx *gorm.DB) error {
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
}
