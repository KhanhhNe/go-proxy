package main

import (
	"embed"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := application.New(application.Options{
		Name: "go-proxy",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})
	app.RegisterService(application.NewService(NewMyService(app)))

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "go-proxy",
		Width:  1224,
		Height: 768,
	})

	err := app.Run()
	if err != nil {
		panic(err)
	}

	if err != nil {
		println("Error:", err.Error())
	}
}
