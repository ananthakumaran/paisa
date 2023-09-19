package price

import "gorm.io/gorm"

type AutoCompleteItem struct {
	Label string `json:"label"`
	ID    string `json:"id"`
}

type AutoCompleteField struct {
	Label     string `json:"label"`
	ID        string `json:"id"`
	Help      string `json:"help"`
	InputType string `json:"inputType"`
}

type PriceProvider interface {
	Code() string
	Label() string
	Description() string
	AutoCompleteFields() []AutoCompleteField
	AutoComplete(db *gorm.DB, field string, filter map[string]string) []AutoCompleteItem
	ClearCache(db *gorm.DB)
}
