package scheme

import (
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

func GetAMCs(db *gorm.DB) []string {
	var amcs []string
	db.Model(&Scheme{}).Distinct().Pluck("AMC", &amcs)
	return amcs
}

func GetNAVNames(db *gorm.DB, amc string) []string {
	var navNames []string
	db.Model(&Scheme{}).Where("amc = ? and type = 'Open Ended'", amc).Pluck("NAVName", &navNames)
	return navNames
}

func FindScheme(db *gorm.DB, amc string, NAVName string) Scheme {
	var scheme Scheme
	result := db.Where("amc = ? and nav_name = ?", amc, NAVName).First(&scheme)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	return scheme
}
