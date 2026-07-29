package main

import (
	"encoding/json"
	"embed"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed wails.json
var wailsConfigFS embed.FS

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

// Version returns the current application version, read from wails.json so it
// stays in sync with the packaged binary version.
func (a *App) Version() string {
	const fallback = "0.1.1"
	data, err := wailsConfigFS.ReadFile("wails.json")
	if err != nil {
		return fallback
	}
	var cfg struct {
		Info struct {
			ProductVersion string `json:"productVersion"`
		} `json:"info"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil || cfg.Info.ProductVersion == "" {
		return fallback
	}
	return cfg.Info.ProductVersion
}
