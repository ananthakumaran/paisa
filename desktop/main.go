package main

import (
	"github.com/ananthakumaran/paisa/cmd"
	"github.com/ananthakumaran/paisa/desktop/logger"
	"github.com/ananthakumaran/paisa/internal/server"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

func main() {
	decimal.MarshalJSONWithoutQuotes = true

	app := NewApp()

	cmd.InitLogger(true, &logger.Hook{
		Ctx: &app.ctx,
		LogLevels: []log.Level{
			log.PanicLevel,
			log.FatalLevel,
		},
	})
	err := wails.Run(&options.App{
		Title:  "Paisa",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Handler: server.Build(&app.db).Handler(),
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Logger: &logger.Logger{},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
