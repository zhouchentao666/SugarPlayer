package main

import (
	_ "embed"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed build/windows/icon.ico
var trayIconBytes []byte

// EnableTray creates or destroys the system tray icon based on the enabled flag.
func (a *App) EnableTray(enabled bool) error {
	if enabled {
		if a.tray != nil {
			return nil
		}
		return a.createTray()
	}
	if a.tray != nil {
		a.tray.Destroy()
		a.tray = nil
		a.traySongLabel = nil
	}
	return nil
}

// SetTraySongInfo updates the disabled "current song" label in the tray menu.
func (a *App) SetTraySongInfo(label string) error {
	if a.traySongLabel != nil {
		a.traySongLabel.SetLabel(label)
	}
	return nil
}

func (a *App) createTray() error {
	tray := a.app.SystemTray.New()
	tray.SetIcon(a.trayIcon)
	tray.SetTooltip("SugarMusic")

	menu := application.NewMenu()
	a.traySongLabel = menu.Add("未在播放")
	a.traySongLabel.SetEnabled(false)
	menu.AddSeparator()
	menu.Add("上一首").OnClick(func(_ *application.Context) {
		a.app.Event.Emit("tray:prev")
	})
	menu.Add("下一首").OnClick(func(_ *application.Context) {
		a.app.Event.Emit("tray:next")
	})
	menu.AddSeparator()
	menu.Add("退出").OnClick(func(_ *application.Context) {
		a.app.Event.Emit("tray:exit")
	})
	tray.SetMenu(menu)

	tray.OnClick(func() {
		a.showMainWindow()
	})

	tray.Run()
	a.tray = tray
	return nil
}

func (a *App) showMainWindow() {
	if a.mainWindow == nil {
		return
	}
	a.mainWindow.Show().Focus()
}
