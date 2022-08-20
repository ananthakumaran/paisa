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

	return gin.H{"expenses": expenses}
}
