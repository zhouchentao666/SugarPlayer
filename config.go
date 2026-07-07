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

// ConfigPlaylistSort stores sort settings for a single playlist.
type ConfigPlaylistSort struct {
	Mode  string `json:"mode"`
	Order string `json:"order"`
}

// ConfigLocalMetadata stores user-edited metadata overrides for a song.
type ConfigLocalMetadata struct {
	Title        string `json:"title"`
	Artist       string `json:"artist"`
	Album        string `json:"album"`
	Cover        string `json:"cover"`
	Lyrics       string `json:"lyrics"`
	LyricsFormat string `json:"lyricsFormat"`
}

// ConfigDesktopLyric stores desktop lyric window and style settings.
type ConfigDesktopLyric struct {
	Enabled             bool   `json:"enabled"`
	FontSize            int    `json:"fontSize"`
	MainColor           string `json:"mainColor"`
	UnplayedColor       string `json:"unplayedColor"`
	ShadowColor         string `json:"shadowColor"`
	FontWeight          int    `json:"fontWeight"`
	Position            string `json:"position"`
	AlwaysShowPlayInfo  bool   `json:"alwaysShowPlayInfo"`
	Animation           bool   `json:"animation"`
	ShowYrc             bool   `json:"showYrc"`
	ShowTran            bool   `json:"showTran"`
	IsDoubleLine        bool   `json:"isDoubleLine"`
	TextBackgroundMask  bool   `json:"textBackgroundMask"`
	BackgroundMaskColor string `json:"backgroundMaskColor"`
	FontFamily          string `json:"fontFamily"`
	X                   int    `json:"x"`
	Y                   int    `json:"y"`
	Width               int    `json:"width"`
	Height              int    `json:"height"`
	IsLock              bool   `json:"isLock"`
}

// ConfigSettings represents app settings in persisted config.
type ConfigSettings struct {
	Theme                string                         `json:"theme"`
	AccentColor          string                         `json:"accentColor"`
	Quality              string                         `json:"quality"`
	Autoplay             bool                           `json:"autoplay"`
	SavePlaylistAndSong  bool                           `json:"savePlaylistAndSong"`
	SaveWindowPosition   bool                           `json:"saveWindowPosition"`
	WindowEffect         string                         `json:"windowEffect"`
	CustomImagePath      string                         `json:"customImagePath"`
	CustomImageOpacity   float64                        `json:"customImageOpacity"`
	CustomImageBlur      float64                        `json:"customImageBlur"`
	SongColorOpacity     float64                        `json:"songColorOpacity"`
	SongColorBlur        float64                        `json:"songColorBlur"`
	FullScreenBackground string                         `json:"fullScreenBackground"`
	ImmersivePlayerBar   bool                           `json:"immersivePlayerBar"`
	Hotkeys              map[string]string              `json:"hotkeys"`
	CheckUpdateOnStartup bool                           `json:"checkUpdateOnStartup"`
	AutoStart            bool                           `json:"autoStart"`
	TrayEnabled          bool                           `json:"trayEnabled"`
	CloseToTray          bool                           `json:"closeToTray"`
	DesktopLyric         ConfigDesktopLyric             `json:"desktopLyric"`
	SelectedPlaylistID   string                         `json:"selectedPlaylistId"`
	PlaylistSorts        map[string]ConfigPlaylistSort  `json:"playlistSorts"`
	LocalMetadata        map[string]ConfigLocalMetadata `json:"localMetadata"`
}

// ConfigPlayback represents the last playback state.
type ConfigPlayback struct {
	PlaylistID string  `json:"playlistId"`
	SongIndex  int     `json:"songIndex"`
	Time       float64 `json:"time"`
}

// ConfigWindow represents the last window bounds.
type ConfigWindow struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// AppConfig represents the persisted application config.
type AppConfig struct {
	Playlists []ConfigPlaylist `json:"playlists"`
	Settings  ConfigSettings   `json:"settings"`
	Playback  ConfigPlayback   `json:"playback"`
	Window    ConfigWindow     `json:"window"`
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
