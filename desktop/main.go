package main

import (
	_ "embed"

	"github.com/ananthakumaran/paisa/cmd"
	"github.com/ananthakumaran/paisa/desktop/logger"
	"github.com/ananthakumaran/paisa/internal/server"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed build/appicon.png
var icon []byte

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
		Title: "Paisa",
		AssetServer: &assetserver.Options{
			Handler: server.Build(&app.db, false).Handler(),
		},
		BackgroundColour: &options.RGBA{R: 250, G: 250, B: 250, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		WindowStartState:         options.Maximised,
		EnableDefaultContextMenu: true,
		Logger:                   &logger.Logger{},
		Mac: &mac.Options{
			About: &mac.AboutInfo{
				Title:   "Paisa",
				Message: "Version 0.5.6 \nCopyright © 2022 - 2023 \nAnantha Kumaran",
				Icon:    icon,
			},
		},

		Linux: &linux.Options{
			Icon:        icon,
			ProgramName: "Paisa",
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
