package goal

import (
	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type GoalSummary struct {
	Type       string          `json:"type"`
	Name       string          `json:"name"`
	Icon       string          `json:"icon"`
	Current    decimal.Decimal `json:"current"`
	Target     decimal.Decimal `json:"target"`
	TargetDate string          `json:"targetDate"`
}

func GetGoalSummaries(db *gorm.DB) []GoalSummary {
	summaries := []GoalSummary{}

	for _, goal := range config.GetConfig().Goals.Retirement {
		summaries = append(summaries, getRetirementSummary(db, goal))
	}

	for _, goal := range config.GetConfig().Goals.Savings {
		summaries = append(summaries, getSavingsSummary(db, goal))
	}

	return summaries
}

func GetGoalDetails(db *gorm.DB, goalType string, name string) gin.H {
	switch goalType {
	case "retirement":
		conf, _ := lo.Find(config.GetConfig().Goals.Retirement, func(conf config.RetirementGoal) bool { return conf.Name == name })
		return getRetirementDetail(db, conf)
	case "savings":
		conf, _ := lo.Find(config.GetConfig().Goals.Savings, func(conf config.SavingsGoal) bool { return conf.Name == name })
		return getSavingsDetail(db, conf)
	}
	return gin.H{}
}
