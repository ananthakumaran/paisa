package portfolio

import (
	"github.com/ananthakumaran/paisa/internal/model/price"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Portfolio struct {
	ID                uint                `gorm:"primaryKey" json:"id"`
	CommodityType     price.CommodityType `json:"commodity_type"`
	ParentCommodityID string              `json:"parent_commodity_id"`
	CommodityID       string              `json:"commodity_id"`
	CommodityName     string              `json:"commodity_name"`
	Percentage        float64             `json:"percentage"`
}

func UpsertAll(db *gorm.DB, commodityType price.CommodityType, parentCommodityID string, portfolios []*Portfolio) {
	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Delete(&Portfolio{}, "commodity_type = ? and parent_commodity_id = ?", commodityType, parentCommodityID).Error
		if err != nil {
			return err
		}
		for _, portfolio := range portfolios {
			err := tx.Create(portfolio).Error
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

func GetPortfolios(db *gorm.DB, parentCommodityID string) []Portfolio {
	var portfolios []Portfolio
	result := db.Model(&Portfolio{}).Where("parent_commodity_id = ?", parentCommodityID).Find(&portfolios)

	if result.Error != nil {
		log.Fatal(result.Error)
	}
	return portfolios
}
