package main

import (
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// FolderWatcher watches music folders and notifies the frontend of changes.
type FolderWatcher struct {
	app     *application.App
	watcher *fsnotify.Watcher
}

func newFolderWatcher(app *application.App) *FolderWatcher {
	return &FolderWatcher{app: app}
}

// WatchMusicFolder watches a folder and its subfolders for changes.
// When files are added or removed, a "folder:changed" event is emitted.
func (a *App) WatchMusicFolder(path string) error {
	return a.watcher.watch(path)
}

// StopWatching stops watching the music folder.
func (a *App) StopWatching() error {
	return a.watcher.stop()
}

func (w *FolderWatcher) watch(path string) error {
	if w.watcher != nil {
		_ = w.watcher.Close()
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	w.watcher = watcher

	err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(p)
		}
		return nil
	})
	if err != nil {
		return err
	}

	go w.loop(path)
	return nil
}

func (w *FolderWatcher) loop(path string) {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Remove == fsnotify.Remove {
				w.app.Event.Emit("folder:changed", path)
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				info, err := os.Stat(event.Name)
				if err == nil && info.IsDir() {
					_ = w.watcher.Add(event.Name)
				}
			}
		case _, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
		}
	}
}

func (w *FolderWatcher) stop() error {
	if w.watcher != nil {
		return w.watcher.Close()
	}
	return nil
}
