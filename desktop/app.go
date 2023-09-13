package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gorm.io/gorm"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
)

// App struct
type App struct {
	ctx context.Context
	db  gorm.DB
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	runtime.WindowMaximise(ctx)

	db, err := gorm.Open(sqlite.Open(config.GetConfig().DBPath), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	model.AutoMigrate(db)

	a.db = *db
	a.ctx = ctx
}
