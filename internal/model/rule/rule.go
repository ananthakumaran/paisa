package rule

import (
	"time"

	"gorm.io/gorm"
)

// Rule represents a regex rule for account prediction
type Rule struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Pattern   string    `json:"pattern" gorm:"not null"`
	Account   string    `json:"account" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName overrides the table name
func (Rule) TableName() string {
	return "rules"
}

// Create creates a new rule
func Create(db *gorm.DB, rule *Rule) error {
	return db.Create(rule).Error
}

// FindAll returns all rules
func FindAll(db *gorm.DB) ([]Rule, error) {
	var rules []Rule
	err := db.Find(&rules).Error
	return rules, err
}

// FindByID returns a rule by ID
func FindByID(db *gorm.DB, id uint) (Rule, error) {
	var rule Rule
	err := db.First(&rule, id).Error
	return rule, err
}

// Update updates a rule
func Update(db *gorm.DB, rule *Rule) error {
	return db.Save(rule).Error
}

// Delete deletes a rule
func Delete(db *gorm.DB, id uint) error {
	return db.Delete(&Rule{}, id).Error
}