package price

import (
	"time"

	"gorm.io/gorm"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/google/btree"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type Price struct {
	ID            uint                 `gorm:"primaryKey" json:"id"`
	Date          time.Time            `json:"date"`
	CommodityType config.CommodityType `json:"commodity_type"`
	CommodityID   string               `json:"commodity_id"`
	CommodityName string               `json:"commodity_name"`
	Value         decimal.Decimal      `json:"value"`
}

func (p Price) Less(o btree.Item) bool {
	return p.Date.Before(o.(Price).Date)
}

func UpsertAllByTypeAndID(db *gorm.DB, commodityType config.CommodityType, commodityID string, prices []*Price) {
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

func UpsertAllByType(db *gorm.DB, commodityType config.CommodityType, prices []Price) {
	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Delete(&Price{}, "commodity_type = ?", commodityType).Error
		if err != nil {
			return err
		}
		for _, price := range prices {
			err := tx.Create(&price).Error
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
