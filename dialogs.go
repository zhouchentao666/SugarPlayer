package main

import (
	"encoding/base64"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// OpenMusicFiles opens a file dialog for selecting multiple music files.
func (a *App) OpenMusicFiles() ([]string, error) {
	return a.app.Dialog.OpenFile().
		SetTitle("选择音乐文件").
		AddFilter("音乐文件", "*.mp3;*.flac;*.wav;*.aac;*.ogg;*.m4a;*.wma;*.opus").
		AddFilter("所有文件", "*.*").
		PromptForMultipleSelection()
}

// OpenMusicFolder opens a directory dialog for selecting a music folder.
func (a *App) OpenMusicFolder() (string, error) {
	return a.app.Dialog.OpenFile().
		CanChooseDirectories(true).
		CanChooseFiles(false).
		SetTitle("选择音乐文件夹").
		PromptForSingleSelection()
}

// OpenImageFile opens a file dialog for selecting a background image.
func (a *App) OpenImageFile() (string, error) {
	return a.app.Dialog.OpenFile().
		SetTitle("选择背景图片").
		AddFilter("图片文件", "*.png;*.jpg;*.jpeg;*.webp;*.bmp").
		AddFilter("所有文件", "*.*").
		PromptForSingleSelection()
}

// ReadImageFile reads an image file and returns it as a base64 data URL.
func (a *App) ReadImageFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	mime := http.DetectContentType(data)
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data), nil
}

// ScanMusicFolder scans a folder recursively for supported music files.
func (a *App) ScanMusicFolder(path string) ([]string, error) {
	var files []string
	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(p))
		switch ext {
		case ".mp3", ".flac", ".wav", ".aac", ".ogg", ".m4a", ".wma", ".opus":
			files = append(files, p)
		}
		return nil
	})
	return files, err
}
