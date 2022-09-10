package server

import (
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetExpense(db *gorm.DB) gin.H {
	expenses := query.Init(db).Like("Expenses:%").NotLike("Expenses:Tax").All()
	incomes := query.Init(db).Like("Income:%").All()
	investments := query.Init(db).Like("Assets:%").NotLike("Assets:Checking").All()
	taxes := query.Init(db).Like("Expenses:Tax").All()

	return gin.H{"expenses": expenses, "month_wise": gin.H{"expenses": posting.GroupByMonth(expenses), "incomes": posting.GroupByMonth(incomes), "investments": posting.GroupByMonth(investments), "taxes": posting.GroupByMonth(taxes)}}
}
