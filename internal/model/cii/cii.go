package cii

import (
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type CII struct {
	ID                 uint   `gorm:"primaryKey" json:"id"`
	FinancialYear      string `json:"financial_year"`
	CostInflationIndex uint   `json:"cost_inflation_index"`
}

func UpsertAll(db *gorm.DB, ciis []*CII) {
	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Exec("DELETE FROM ciis").Error
		if err != nil {
			return err
		}
		for _, cii := range ciis {
			err := tx.Create(cii).Error
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

func GetIndex(db *gorm.DB, financialYear string) uint {
	var cii CII
	result := db.Where("financial_year = ?", financialYear).First(&cii)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	return cii.CostInflationIndex
}
