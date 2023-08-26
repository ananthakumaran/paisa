package main

import (
	"html/template"

	"github.com/ananthakumaran/paisa/cmd"
	"github.com/ananthakumaran/paisa/internal/config"
	mutualfund "github.com/ananthakumaran/paisa/internal/model/mutualfund/scheme"
	nps "github.com/ananthakumaran/paisa/internal/model/nps/scheme"
	"github.com/ananthakumaran/paisa/internal/model/portfolio"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/server"
	"github.com/shopspring/decimal"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	decimal.MarshalJSONWithoutQuotes = true
	cmd.InitConfig()

	db, err := gorm.Open(sqlite.Open(config.GetConfig().DBPath), &gorm.Config{})
	db.AutoMigrate(&nps.Scheme{})
	db.AutoMigrate(&mutualfund.Scheme{})
	db.AutoMigrate(&posting.Posting{})
	db.AutoMigrate(&price.Price{})
	db.AutoMigrate(&portfolio.Portfolio{})
	db.AutoMigrate(&template.Template{})

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "desktop",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Handler: server.Build(db).Handler(),
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Debug: options.Debug{
			OpenInspectorOnStartup: true,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
