package main

import (
	"QueryRaccoon/bindings"
	"QueryRaccoon/internal/connections"
	"QueryRaccoon/internal/query"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	manager := connections.NewManager()
	queryService := query.NewService(manager)

	connBinding := bindings.NewConnectionService(manager)
	queryBinding := bindings.NewQueryService(queryService)

	err := wails.Run(&options.App{
		Title:  "QueryRaccoon",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
			connBinding,
			queryBinding,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
