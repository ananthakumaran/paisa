package server

import (
	"github.com/ananthakumaran/paisa/internal/model/posting"
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

	var incomes []posting.Posting
	result = db.Where("account like ? order by date asc", "Income:%").Find(&incomes)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	var investments []posting.Posting
	result = db.Where("account like ? and account != ? order by date asc", "Assets:%", "Assets:Checking").Find(&investments)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	var taxes []posting.Posting
	result = db.Where("account = ? order by date asc", "Expenses:Tax").Find(&taxes)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	return gin.H{"expenses": expenses, "month_wise": gin.H{"expenses": posting.GroupByMonth(expenses), "incomes": posting.GroupByMonth(incomes), "investments": posting.GroupByMonth(investments), "taxes": posting.GroupByMonth(taxes)}}
}
