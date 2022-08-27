package server

import (
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func GetExpense(db *gorm.DB) gin.H {
	var expenses []posting.Posting
	result := db.Where("account like ? and account != ? order by date asc", "Expenses:%", "Expenses:Tax").Find(&expenses)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	now := time.Now()
	start := utils.BeginningOfMonth(now)
	end := utils.EndOfMonth(now)

	var currentExpenses []posting.Posting
	result = db.Debug().Where("account like ? and account != ? and date >= ? and date <= ? order by date asc", "Expenses:%", "Expenses:Tax", start, end).Find(&currentExpenses)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	var currentIncomes []posting.Posting
	result = db.Where("account like ? and date >= ? and date <= ? order by date asc", "Income:%", start, end).Find(&currentIncomes)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	var currentInvestments []posting.Posting
	result = db.Where("account like ? and account != ? and date >= ? and date <= ? order by date asc", "Assets:%", "Assets:Checking", start, end).Find(&currentInvestments)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	var currentTaxes []posting.Posting
	result = db.Where("account = ? and date >= ? and date <= ? order by date asc", "Expenses:Tax", start, end).Find(&currentTaxes)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	return gin.H{"expenses": expenses, "current_month": gin.H{"expenses": currentExpenses, "incomes": currentIncomes, "investments": currentInvestments, "taxes": currentTaxes}}
}
