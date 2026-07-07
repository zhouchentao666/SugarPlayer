package main

import (
	_ "embed"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed build/windows/icon.ico
var trayIconBytes []byte

// EnableTray creates or destroys the system tray icon based on the enabled flag.
func (a *App) EnableTray(enabled bool) error {
	a.trayMu.Lock()
	defer a.trayMu.Unlock()

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
	a.trayMu.Lock()
	defer a.trayMu.Unlock()

	if a.traySongLabel != nil {
		a.traySongLabel.SetLabel(label)
	}
	return nil
}

func (a *App) createTray() error {
	tray := a.app.SystemTray.New()
	a.tray = tray
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
	a.trayLyricToggleLabel = menu.Add("显示桌面歌词")
	a.trayLyricToggleLabel.OnClick(func(_ *application.Context) {
		a.toggleDesktopLyricFromTray()
	})
	a.trayLyricLockLabel = menu.Add("锁定桌面歌词")
	a.trayLyricLockLabel.OnClick(func(_ *application.Context) {
		a.toggleDesktopLyricLockFromTray()
	})
	menu.AddSeparator()
	menu.Add("主界面").OnClick(func(_ *application.Context) {
		a.showMainWindow()
	})
	menu.AddSeparator()
	menu.Add("退出").OnClick(func(_ *application.Context) {
		a.app.Event.Emit("tray:exit")
	})
	tray.SetMenu(menu)
	a.updateTrayLyricMenu()

	tray.OnClick(func() {
		a.showMainWindow()
	})

	return nil
}

func (a *App) toggleDesktopLyricFromTray() {
	cfg := a.loadDesktopLyricConfig()
	next := !cfg.Enabled
	_ = a.ToggleDesktopLyric(next)
	a.updateTrayLyricMenu()
}

func (a *App) toggleDesktopLyricLockFromTray() {
	cfg := a.loadDesktopLyricConfig()
	next := !cfg.IsLock
	_ = a.SetDesktopLyricIgnoreMouseEvents(next)
	a.updateTrayLyricMenu()
}

func (a *App) updateTrayLyricMenu() {
	if a.tray == nil {
		return
	}
	cfg := a.loadDesktopLyricConfig()
	if a.trayLyricToggleLabel != nil {
		if cfg.Enabled {
			a.trayLyricToggleLabel.SetLabel("隐藏桌面歌词")
		} else {
			a.trayLyricToggleLabel.SetLabel("显示桌面歌词")
		}
	}
	if a.trayLyricLockLabel != nil {
		if cfg.IsLock {
			a.trayLyricLockLabel.SetLabel("解锁桌面歌词")
		} else {
			a.trayLyricLockLabel.SetLabel("锁定桌面歌词")
		}
	}
}

func (a *App) showMainWindow() {
	if a.mainWindow == nil {
		return
	}
	a.mainWindow.Show().Focus()
}
