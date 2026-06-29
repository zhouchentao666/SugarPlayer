package main

import (
	"context"
	"fmt"

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
