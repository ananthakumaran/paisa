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

func DeleteAll(db *gorm.DB) error {
	err := db.Exec("DELETE FROM prices").Error
	if err != nil {
		return err
	}
	return nil
}

func UpsertAllByTypeAndName(db *gorm.DB, commodityType config.CommodityType, commodityName string, prices []*Price) error {
	return db.Transaction(func(tx *gorm.DB) error {
		previous := tx.Where("commodity_name = ?", commodityName)
		if commodityType == config.Unknown {
			previous = previous.Where("commodity_type = ?", config.Unknown)
		} else {
			previous = previous.Where("commodity_type != ?", config.Unknown)
		}
		err := previous.Delete(&Price{}).Error
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
