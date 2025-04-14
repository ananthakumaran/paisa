package stock_tag

import (
	"gorm.io/gorm"
)

type StockTag struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Symbol string `gorm:"index" json:"symbol"`
	Tag    string `gorm:"index" json:"tag"`
	Color  string `json:"color"`
}

func (StockTag) TableName() string {
	return "stock_tags"
}

func GetTags(db *gorm.DB, symbol string) ([]StockTag, error) {
	var tags []StockTag
	result := db.Where("symbol = ?", symbol).Find(&tags)
	return tags, result.Error
}

func AddTag(db *gorm.DB, symbol string, tag string, color string) error {
	return db.Create(&StockTag{
		Symbol: symbol,
		Tag:    tag,
		Color:  color,
	}).Error
}

func RemoveTag(db *gorm.DB, symbol string, tag string) error {
	return db.Where("symbol = ? AND tag = ?", symbol, tag).Delete(&StockTag{}).Error
}

func GetAllTags(db *gorm.DB) (map[string][]StockTag, error) {
	var tags []StockTag
	result := db.Find(&tags)
	if result.Error != nil {
		return nil, result.Error
	}

	tagMap := make(map[string][]StockTag)
	for _, tag := range tags {
		tagMap[tag.Symbol] = append(tagMap[tag.Symbol], tag)
	}
	return tagMap, nil
}
