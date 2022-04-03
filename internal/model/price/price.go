package price

import (
	"gorm.io/gorm"
	"time"

	log "github.com/sirupsen/logrus"
)

type CommodityType string

const (
	MutualFund CommodityType = "mutualfund"
)

type Price struct {
	ID            uint `gorm:"primaryKey"`
	Date          time.Time
	CommodityType CommodityType
	CommodityID   string
	Value         float64
}

func UpsertAll(db *gorm.DB, commodityType CommodityType, commodityID string, prices []*Price) {
	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Delete(&Price{}, "commodity_type = ? and commodity_id = ?", commodityType, commodityID).Error
		if err != nil {
			return err
		}
		for _, price := range prices {
			err := tx.Create(price).Error
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
