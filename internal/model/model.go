package model

import (
	"github.com/ananthakumaran/paisa/internal/ledger"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func Sync(db *gorm.DB) {
	db.AutoMigrate(&posting.Posting{})
	log.Info("Syncing postings from ledger")
	postings, _ := ledger.Parse(viper.GetString("journal_path"))
	posting.UpsertAll(db, postings)
}
