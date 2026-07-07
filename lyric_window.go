package main

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

const desktopLyricWindowName = "desktop-lyric"

func (a *App) loadDesktopLyricConfig() ConfigDesktopLyric {
	cfg, err := a.LoadConfig()
	if err != nil {
		return defaultDesktopLyricConfig()
	}
	return mergeDesktopLyricConfig(cfg.Settings.DesktopLyric)
}

func defaultDesktopLyricConfig() ConfigDesktopLyric {
	return ConfigDesktopLyric{
		FontSize:            30,
		MainColor:           "#73BCFC",
		UnplayedColor:       "rgba(255, 255, 255, 0.5)",
		ShadowColor:         "rgba(255, 255, 255, 0.5)",
		FontWeight:          600,
		Position:            "center",
		AlwaysShowPlayInfo:  false,
		Animation:           true,
		ShowYrc:             true,
		ShowTran:            false,
		IsDoubleLine:        true,
		TextBackgroundMask:  false,
		BackgroundMaskColor: "rgba(0,0,0,0.2)",
		FontFamily:          "PingFangSC-Semibold, system-ui, -apple-system, sans-serif",
		Width:               800,
		Height:              180,
		IsLock:              false,
	}
}

func mergeDesktopLyricConfig(cfg ConfigDesktopLyric) ConfigDesktopLyric {
	def := defaultDesktopLyricConfig()
	if cfg.FontSize > 0 {
		def.FontSize = cfg.FontSize
	}
	if cfg.MainColor != "" {
		def.MainColor = cfg.MainColor
	}
	if cfg.UnplayedColor != "" {
		def.UnplayedColor = cfg.UnplayedColor
	}
	if cfg.ShadowColor != "" {
		def.ShadowColor = cfg.ShadowColor
	}
	if cfg.FontWeight > 0 {
		def.FontWeight = cfg.FontWeight
	}
	if cfg.Position != "" {
		def.Position = cfg.Position
	}
	if cfg.BackgroundMaskColor != "" {
		def.BackgroundMaskColor = cfg.BackgroundMaskColor
	}
	if cfg.FontFamily != "" {
		def.FontFamily = cfg.FontFamily
	}
	if cfg.Width > 0 {
		def.Width = cfg.Width
	}
	if cfg.Height > 0 {
		def.Height = cfg.Height
	}
	def.Enabled = cfg.Enabled
	def.AlwaysShowPlayInfo = cfg.AlwaysShowPlayInfo
	def.Animation = cfg.Animation
	def.ShowYrc = cfg.ShowYrc
	def.ShowTran = cfg.ShowTran
	def.IsDoubleLine = cfg.IsDoubleLine
	def.TextBackgroundMask = cfg.TextBackgroundMask
	def.IsLock = cfg.IsLock
	def.X = cfg.X
	def.Y = cfg.Y
	return def
}

func (a *App) ensureDesktopLyricPosition(cfg *ConfigDesktopLyric) {
	primary := a.app.Screen.GetPrimary()
	if primary == nil {
		if cfg.X == 0 && cfg.Y == 0 {
			cfg.X = 100
			cfg.Y = 100
		}
		return
	}
	work := primary.WorkArea
	if cfg.Width <= 0 {
		cfg.Width = 800
	}
	if cfg.Height <= 0 {
		cfg.Height = 180
	}
	if cfg.X == 0 && cfg.Y == 0 {
		cfg.X = work.X + work.Width/2 - cfg.Width/2
		cfg.Y = work.Y + work.Height - cfg.Height - 90
	}
	if cfg.X+cfg.Width > work.X+work.Width {
		cfg.X = work.X + work.Width - cfg.Width
	}
	if cfg.Y+cfg.Height > work.Y+work.Height {
		cfg.Y = work.Y + work.Height - cfg.Height
	}
	if cfg.X < work.X {
		cfg.X = work.X
	}
	if cfg.Y < work.Y {
		cfg.Y = work.Y
	}
}

// ToggleDesktopLyric shows or hides the desktop lyric window.
func (a *App) ToggleDesktopLyric(enabled bool) error {
	a.desktopLyricWindowMu.Lock()
	defer a.desktopLyricWindowMu.Unlock()

	if enabled {
		if a.desktopLyricWindow != nil {
			a.desktopLyricWindow.Show().SetAlwaysOnTop(true)
			return nil
		}
		cfg := a.loadDesktopLyricConfig()
		a.ensureDesktopLyricPosition(&cfg)
		win := a.app.Window.NewWithOptions(application.WebviewWindowOptions{
			Name:            desktopLyricWindowName,
			Title:           "桌面歌词",
			Width:           cfg.Width,
			Height:          cfg.Height,
			X:               cfg.X,
			Y:               cfg.Y,
			Frameless:       true,
			AlwaysOnTop:     true,
			DisableResize:   false,
			MinWidth:        440,
			MinHeight:       120,
			MaxWidth:        1600,
			MaxHeight:       300,
			Hidden:          true,
			BackgroundType:  application.BackgroundTypeTransparent,
			URL:             "/?desktopLyric=1",
			DefaultContextMenuDisabled: true,
			Windows: application.WindowsWindow{
				DisableFramelessWindowDecorations: true,
				HiddenOnTaskbar:                   true,
			},
		})
		if win == nil {
			return nil
		}
		a.desktopLyricWindow = win
		a.bindDesktopLyricWindowEvents(win)
		win.SetIgnoreMouseEvents(cfg.IsLock)
		win.Show().SetAlwaysOnTop(true)
		return nil
	}

	if a.desktopLyricWindow != nil {
		a.desktopLyricWindow.Hide()
	}
	cfg := a.loadDesktopLyricConfig()
	cfg.Enabled = enabled
	a.saveDesktopLyricConfig(cfg)
	a.updateTrayLyricMenu()
	return nil
}

func (a *App) bindDesktopLyricWindowEvents(win application.Window) {
	var mu sync.Mutex
	var debounceTimer *time.Timer
	const debounceMs = 200

	saveBounds := func() {
		cfg := a.loadDesktopLyricConfig()
		bounds := win.Bounds()
		cfg.X = bounds.X
		cfg.Y = bounds.Y
		cfg.Width = bounds.Width
		cfg.Height = bounds.Height
		a.saveDesktopLyricConfig(cfg)
	}

	debouncedSave := func() {
		mu.Lock()
		if debounceTimer != nil {
			debounceTimer.Stop()
		}
		debounceTimer = time.AfterFunc(debounceMs, func() {
			saveBounds()
		})
		mu.Unlock()
	}

	win.RegisterHook(events.Common.WindowDidMove, func(_ *application.WindowEvent) {
		debouncedSave()
	})
	win.RegisterHook(events.Common.WindowDidResize, func(_ *application.WindowEvent) {
		debouncedSave()
	})
}

func (a *App) saveDesktopLyricConfig(cfg ConfigDesktopLyric) {
	appCfg, err := a.LoadConfig()
	if err != nil {
		appCfg = AppConfig{}
	}
	appCfg.Settings.DesktopLyric = cfg
	_ = a.SaveConfig(appCfg)
}

// SetDesktopLyricBounds updates the desktop lyric window bounds and persists them.
func (a *App) SetDesktopLyricBounds(x, y, width, height int) error {
	a.desktopLyricWindowMu.Lock()
	defer a.desktopLyricWindowMu.Unlock()

	if a.desktopLyricWindow == nil {
		cfg := a.loadDesktopLyricConfig()
		cfg.X = x
		cfg.Y = y
		cfg.Width = width
		cfg.Height = height
		a.saveDesktopLyricConfig(cfg)
		return nil
	}
	a.desktopLyricWindow.SetBounds(application.Rect{X: x, Y: y, Width: width, Height: height})
	cfg := a.loadDesktopLyricConfig()
	cfg.X = x
	cfg.Y = y
	cfg.Width = width
	cfg.Height = height
	a.saveDesktopLyricConfig(cfg)
	return nil
}

// SetDesktopLyricIgnoreMouseEvents toggles mouse event ignoring for the lyric window.
func (a *App) SetDesktopLyricIgnoreMouseEvents(ignore bool) error {
	a.desktopLyricWindowMu.Lock()
	defer a.desktopLyricWindowMu.Unlock()

	if a.desktopLyricWindow != nil {
		a.desktopLyricWindow.SetIgnoreMouseEvents(ignore)
	}
	cfg := a.loadDesktopLyricConfig()
	cfg.IsLock = ignore
	a.saveDesktopLyricConfig(cfg)
	a.updateTrayLyricMenu()
	return nil
}

// GetDesktopLyricConfig returns the current desktop lyric configuration as JSON.
func (a *App) GetDesktopLyricConfig() string {
	cfg := a.loadDesktopLyricConfig()
	data, _ := json.Marshal(cfg)
	return string(data)
}

// CloseDesktopLyric closes and destroys the desktop lyric window.
func (a *App) CloseDesktopLyric() error {
	a.desktopLyricWindowMu.Lock()
	defer a.desktopLyricWindowMu.Unlock()

	if a.desktopLyricWindow != nil {
		a.desktopLyricWindow.Close()
		a.desktopLyricWindow = nil
	}
	cfg := a.loadDesktopLyricConfig()
	cfg.Enabled = false
	a.saveDesktopLyricConfig(cfg)
	a.updateTrayLyricMenu()
	return nil
}
