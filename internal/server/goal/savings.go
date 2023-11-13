package goal

import (
	"github.com/ananthakumaran/paisa/internal/accounting"
	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func getSavingsSummary(db *gorm.DB, conf config.SavingsGoal) GoalSummary {
	savings := accounting.FilterByGlob(query.Init(db).Like("Assets:%").All(), conf.Accounts)
	savings = service.PopulateMarketPrice(db, savings)
	savingsTotal := accounting.CurrentBalance(savings)

	return GoalSummary{
		Type:       "savings",
		Name:       conf.Name,
		Current:    savingsTotal,
		Target:     decimal.NewFromFloat(conf.Target),
		TargetDate: conf.TargetDate,
		Icon:       conf.Icon,
	}
}

func getSavingsDetail(db *gorm.DB, conf config.SavingsGoal) gin.H {
	savings := accounting.FilterByGlob(query.Init(db).Like("Assets:%").All(), conf.Accounts)
	savings = service.PopulateMarketPrice(db, savings)
	savingsTotal := accounting.CurrentBalance(savings)

	savingsWithCapitalGains := accounting.FilterByGlob(query.Init(db).Like("Assets:%", "Income:CapitalGains:%").All(), conf.Accounts)
	savingsWithCapitalGains = service.PopulateMarketPrice(db, savingsWithCapitalGains)

	return gin.H{
		"type":            "savings",
		"name":            conf.Name,
		"icon":            conf.Icon,
		"savingsTimeline": accounting.RunningBalance(db, savings),
		"savingsTotal":    savingsTotal,
		"target":          decimal.NewFromFloat(conf.Target),
		"targetDate":      conf.TargetDate,
		"rate":            conf.Rate,
		"xirr":            service.XIRR(db, savingsWithCapitalGains),
		"postings":        savingsWithCapitalGains,
	}
}
