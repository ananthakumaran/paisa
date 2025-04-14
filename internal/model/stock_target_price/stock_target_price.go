package stock_target_price

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type StockTargetPrice struct {
	Symbol      string          `gorm:"primaryKey"`
	TargetPrice decimal.Decimal `gorm:"type:decimal(20,8)"`
}

func SetTargetPrice(db *gorm.DB, symbol string, targetPrice decimal.Decimal) error {
	return db.Save(&StockTargetPrice{
		Symbol:      symbol,
		TargetPrice: targetPrice,
	}).Error
}

func GetTargetPrice(db *gorm.DB, symbol string) (decimal.Decimal, error) {
	var targetPrice StockTargetPrice
	err := db.First(&targetPrice, "symbol = ?", symbol).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return decimal.Zero, nil
		}
		return decimal.Zero, err
	}
	return targetPrice.TargetPrice, nil
}
