package scheme

import (
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Scheme struct {
	ID         uint `gorm:"primaryKey" json:"id"`
	PFMName    string
	SchemeID   string
	SchemeName string
}

func (Scheme) TableName() string {
	return "nps_schemes"
}

func Count(db *gorm.DB) int64 {
	var count int64
	db.Model(&Scheme{}).Count(&count)
	return count
}

func UpsertAll(db *gorm.DB, schemes []*Scheme) {
	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Exec("DELETE FROM nps_schemes").Error
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

func GetPFMs(db *gorm.DB) []string {
	var pfms []string
	db.Model(&Scheme{}).Distinct().Pluck("PFMName", &pfms)
	return pfms
}

func GetSchemeNames(db *gorm.DB, pfm string) []string {
	var schemeNames []string
	db.Model(&Scheme{}).Where("pfm_name = ?", pfm).Pluck("SchemeName", &schemeNames)
	return schemeNames
}

func FindScheme(db *gorm.DB, pfm string, schemeName string) Scheme {
	var scheme Scheme
	result := db.Where("pfm_name = ? and scheme_name = ?", pfm, schemeName).First(&scheme)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	return scheme
}
