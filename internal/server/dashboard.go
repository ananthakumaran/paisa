package server

import (
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetDashboard(db *gorm.DB) gin.H {
	return gin.H{
		"networth":             GetCurrentNetworth(db),
		"expenses":             GetCurrentExpense(db),
		"cashFlows":            GetCurrentCashFlow(db),
		"transactionSequences": ComputeRecurringTransactions(query.Init(db).All()),
		"transactions":         GetLatestTransactions(db),
	}
}
