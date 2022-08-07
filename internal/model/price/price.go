package price

import (
	"time"

	"gorm.io/gorm"

	"github.com/google/btree"
	log "github.com/sirupsen/logrus"
)

type CommodityType string

const (
	MutualFund CommodityType = "mutualfund"
	NPS        CommodityType = "nps"
)

type Price struct {
	ID            uint `gorm:"primaryKey"`
	Date          time.Time
	CommodityType CommodityType
	CommodityID   string
	CommodityName string
	Value         float64
}

func (p Price) Less(o btree.Item) bool {
	return p.Date.Before(o.(Price).Date)
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
