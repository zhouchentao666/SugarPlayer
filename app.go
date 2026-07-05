package main

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// App struct
type App struct {
	app     *application.App
	audio   *AudioServer
	watcher *FolderWatcher
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
	return nil
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
