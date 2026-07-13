package main

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// App struct
type App struct {
	app                    *application.App
	audio                  *AudioServer
	watcher                *FolderWatcher
	mainWindow             application.Window
	trayMu                 sync.Mutex
	tray                   *application.SystemTray
	traySongLabel          *application.MenuItem
	trayLyricToggleLabel   *application.MenuItem
	trayLyricLockLabel     *application.MenuItem
	trayIcon               []byte
	closeToTray            bool
	desktopLyricWindow     application.Window
	desktopLyricWindowMu   sync.Mutex
	downloadUnlocked       bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// ServiceStartup is called when the app starts.
func (a *App) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	a.app = application.Get()
	a.audio = newAudioServer()
	a.watcher = newFolderWatcher(a.app)
	a.trayIcon = trayIconBytes
	a.loadDownloadUnlock()

	if win, ok := a.app.Window.GetByName("main"); ok {
		a.mainWindow = win
		win.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
			if a.shouldCloseToTray() {
				event.Cancel()
				win.Hide()
			}
		})
	}
	return nil
}

// SetCloseToTray updates whether the close button should hide the window to tray.
func (a *App) SetCloseToTray(enabled bool) error {
	a.trayMu.Lock()
	defer a.trayMu.Unlock()

	a.closeToTray = enabled
	return nil
}

func (a *App) shouldCloseToTray() bool {
	a.trayMu.Lock()
	defer a.trayMu.Unlock()

	return a.closeToTray && a.tray != nil
}

// ServiceName returns the name of the service.
func (a *App) ServiceName() string {
	return "App"
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// OpenInExplorer opens the file explorer and selects the given path.
func (a *App) OpenInExplorer(path string) error {
	return exec.Command("explorer", "/select,", path).Start()
}

// OpenSongEditor opens a dedicated editor window for the given song path.
func (a *App) OpenSongEditor(path string) error {
	editorURL := "/?editor=1&path=" + url.QueryEscape(path)
	if win, ok := a.app.Window.GetByName("song-editor"); ok {
		win.SetURL(editorURL)
		win.Focus()
		win.Show()
		return nil
	}
	a.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:           "song-editor",
		Title:          "编辑歌曲信息",
		Width:          560,
		Height:         680,
		MinWidth:       400,
		MinHeight:      500,
		Frameless:      false,
		URL:            editorURL,
		BackgroundType: application.BackgroundTypeTranslucent,
		Windows: application.WindowsWindow{
			BackdropType: application.Acrylic,
		},
	})
	return nil
}

// EmitMetadataChanged emits an application-wide event to notify all windows that local metadata has changed.
func (a *App) EmitMetadataChanged() {
	a.app.Event.Emit("localmetadata:changed", nil)
}

// OpenURL opens the given URL in the default system browser.
func (a *App) OpenURL(u string) error {
	return a.app.Browser.OpenURL(u)
}

// ShowMainWindow shows and focuses the main application window.
func (a *App) ShowMainWindow() error {
	a.showMainWindow()
	return nil
}
