package main

import (
	"embed"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := application.New(application.Options{
		Name: "SugarMusic",
		Services: []application.Service{
			application.NewService(&App{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(&assets),
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:           "SugarMusic",
		Width:           800,
		Height:          600,
		Frameless:       true,
		BackgroundType:  application.BackgroundTypeTranslucent,
		Windows: application.WindowsWindow{
			BackdropType: application.Acrylic,
		},
		Mac: application.MacWindow{
			Backdrop:   application.MacBackdropTransparent,
			Appearance: application.DefaultAppearance,
		},
	})

	app.Run()
}
