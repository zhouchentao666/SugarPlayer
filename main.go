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
		Name:           "main",
		Title:          "SugarMusic",
		Width:          800,
		Height:         600,
		Frameless:      true,
		BackgroundType: application.BackgroundTypeTranslucent,
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

// Version returns the current application version.
func (a *App) Version() string {
	return "0.0.7"
}
