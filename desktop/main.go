package main

import (
	"github.com/ananthakumaran/paisa/cmd"
	"github.com/ananthakumaran/paisa/internal/server"
	"github.com/shopspring/decimal"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

func main() {

	decimal.MarshalJSONWithoutQuotes = true

	cmd.InitConfig()
	app := NewApp()
	err := wails.Run(&options.App{
		Title:  "desktop",
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
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
