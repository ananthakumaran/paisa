package scheme

import (
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Scheme struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	AMC      string
	Code     string
	Name     string
	Type     string
	Category string
	NAVName  string
}

func Count(db *gorm.DB) int64 {
	var count int64
	db.Model(&Scheme{}).Count(&count)
	return count
}

func UpsertAll(db *gorm.DB, schemes []*Scheme) {
	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Exec("DELETE FROM schemes").Error
		if err != nil {
			return err
		}
		for _, scheme := range schemes {
			err := tx.Create(scheme).Error
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

func GetAMCCompletions(db *gorm.DB) []price.AutoCompleteItem {
	var amcs []string
	db.Model(&Scheme{}).Distinct().Pluck("AMC", &amcs)
	return lo.Map(amcs, func(amc string, _ int) price.AutoCompleteItem {
		return price.AutoCompleteItem{Label: amc, ID: amc}
	})
}

func GetNAVNameCompletions(db *gorm.DB, amc string) []price.AutoCompleteItem {
	var schemes []Scheme
	db.Model(&Scheme{}).Where("amc = ? and type = 'Open Ended'", amc).Find(&schemes)
	return lo.Map(schemes, func(scheme Scheme, _ int) price.AutoCompleteItem {
		return price.AutoCompleteItem{Label: scheme.NAVName, ID: scheme.Code}
	})
}
