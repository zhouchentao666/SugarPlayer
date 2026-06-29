package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// SongMetadata holds audio metadata for a single file.
type SongMetadata struct {
	Title    string  `json:"title"`
	Artist   string  `json:"artist"`
	Album    string  `json:"album"`
	Genre    string  `json:"genre"`
	Year     string  `json:"year"`
	Duration float64 `json:"duration"`
	Bitrate  uint    `json:"bitrate"`
}

// ConfigSong represents a song in persisted config.
type ConfigSong struct {
	ID       string        `json:"id"`
	Path     string        `json:"path"`
	Title    string        `json:"title"`
	Metadata *SongMetadata `json:"metadata,omitempty"`
}

// ConfigPlaylist represents a playlist in persisted config.
type ConfigPlaylist struct {
	ID      string       `json:"id"`
	Name    string       `json:"name"`
	Songs   []ConfigSong `json:"songs"`
	Folders []string     `json:"folders"`
}

// ConfigSettings represents app settings in persisted config.
type ConfigSettings struct {
	Theme              string  `json:"theme"`
	AccentColor        string  `json:"accentColor"`
	Quality            string  `json:"quality"`
	Autoplay           bool    `json:"autoplay"`
	WindowEffect       string  `json:"windowEffect"`
	CustomImagePath    string  `json:"customImagePath"`
	CustomImageOpacity float64 `json:"customImageOpacity"`
	CustomImageBlur    float64 `json:"customImageBlur"`
	SongColorOpacity   float64 `json:"songColorOpacity"`
	SongColorBlur      float64 `json:"songColorBlur"`
}

// AppConfig represents the persisted application config.
type AppConfig struct {
	Playlists []ConfigPlaylist `json:"playlists"`
	Settings  ConfigSettings   `json:"settings"`
}

func configPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	return filepath.Join(configDir, "SugarMusic", "config.json")
}

// SaveConfig persists the application config to disk.
func (a *App) SaveConfig(config AppConfig) error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadConfig loads the application config from disk.
func (a *App) LoadConfig() (AppConfig, error) {
	path := configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return AppConfig{}, err
	}
	var config AppConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return AppConfig{}, err
	}
	return config, nil
}
