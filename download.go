package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"sugarplayer/internal/music/core"
)

// downloadKey is the one-time key required to unlock the download feature.
// It is provided by the user for personal use only; once verified it is
// persisted locally so subsequent downloads need no key.
const downloadKey = "7sK9pR2vG5bN8dQ4zX0cT1jL6mW3aY7"

// OnlineDownloadOpts controls what gets downloaded for an online song.
type OnlineDownloadOpts struct {
	Dir        string `json:"dir"`
	WithLyrics bool   `json:"withLyrics"`
	WithCover  bool   `json:"withCover"`
	Embed      bool   `json:"embed"`
	Quality    string `json:"quality"`
}

// OnlineDownloadResult reports where the downloaded files were written.
type OnlineDownloadResult struct {
	Path      string `json:"path"`
	LyricPath string `json:"lyricPath"`
	CoverPath string `json:"coverPath"`
	Warning   string `json:"warning"`
}

func downloadUnlockPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	return filepath.Join(configDir, "SugarMusic", "download_unlock")
}

func (a *App) loadDownloadUnlock() {
	data, err := os.ReadFile(downloadUnlockPath())
	if err != nil {
		a.downloadUnlocked = false
		return
	}
	a.downloadUnlocked = strings.TrimSpace(string(data)) == "1"
}

func (a *App) persistDownloadUnlock() {
	path := downloadUnlockPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return
	}
	val := "0"
	if a.downloadUnlocked {
		val = "1"
	}
	_ = os.WriteFile(path, []byte(val), 0644)
}

// OnlineVerifyKey checks the download key. On success it unlocks downloading
// for the current and all future sessions.
func (a *App) OnlineVerifyKey(key string) bool {
	if strings.TrimSpace(key) == downloadKey {
		a.downloadUnlocked = true
		a.persistDownloadUnlock()
		return true
	}
	return false
}

// OnlineIsUnlocked reports whether downloading has been unlocked.
func (a *App) OnlineIsUnlocked() bool {
	return a.downloadUnlocked
}

// OnlineDownload downloads an online song to disk according to the options.
// It requires the download feature to be unlocked via OnlineVerifyKey.
func (a *App) OnlineDownload(song OnlineSong, opts OnlineDownloadOpts) (*OnlineDownloadResult, error) {
	if !a.downloadUnlocked {
		return nil, fmt.Errorf("需要下载密钥")
	}

	ms := toModelSong(song)
	ms.Cover = realCoverURL(ms.Cover)
	// 下载使用当前选择的音质（网易云 / QQ / 酷狗 / 酷我 通过 ZQ 网关生效）
	quality := strings.TrimSpace(opts.Quality)
	if quality != "" && (ms.Source == "netease" || ms.Source == "qq" || ms.Source == "kugou" || ms.Source == "kuwo") {
		if ms.Extra == nil {
			ms.Extra = map[string]string{}
		}
		ms.Extra["quality"] = quality
	}

	outDir := strings.TrimSpace(opts.Dir)
	if outDir == "" {
		if home, err := os.UserHomeDir(); err == nil {
			outDir = filepath.Join(home, "Music", "SugarPlayer")
		} else {
			outDir = "."
		}
	}
	template := core.DefaultDownloadFilenameTemplate

	result := &OnlineDownloadResult{}

	if opts.Embed {
		ds, err := core.SaveSongToFileWithTemplate(ms, outDir, true, true, template)
		if err != nil {
			return nil, err
		}
		result.Path = ds.SavedPath
		result.Warning = ds.Warning
	} else {
		ds, err := core.DownloadSongDataWithTemplate(ms, false, false, template)
		if err != nil {
			return nil, err
		}
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return nil, err
		}
		audioPath := filepath.Join(outDir, ds.Filename)
		if err := os.WriteFile(audioPath, ds.Data, 0644); err != nil {
			return nil, err
		}
		result.Path = audioPath
	}

	base := strings.TrimSuffix(result.Path, filepath.Ext(result.Path))

	if opts.WithLyrics {
		if lf := core.GetLyricFunc(ms.Source); lf != nil {
			if lrc, err := lf(ms); err == nil && strings.TrimSpace(lrc) != "" {
				lp := base + ".lrc"
				if err := os.WriteFile(lp, []byte(lrc), 0644); err == nil {
					result.LyricPath = lp
				}
			}
		}
	}

	if opts.WithCover && strings.TrimSpace(ms.Cover) != "" {
		if data, mime, err := core.FetchBytesWithMime(ms.Cover, ms.Source); err == nil && len(data) > 0 {
			ext := coverExtFromMime(mime)
			if ext == "" {
				ext = ".jpg"
			}
			cp := base + ext
			if err := os.WriteFile(cp, data, 0644); err == nil {
				result.CoverPath = cp
			}
		}
	}

	return result, nil
}

// realCoverURL unwraps the local /cover proxy URL back to the original remote
// cover URL so downloads/embedding fetch the image directly.
func realCoverURL(proxy string) string {
	if strings.Contains(proxy, "/cover?url=") {
		if u, err := url.Parse(proxy); err == nil {
			if r := u.Query().Get("url"); r != "" {
				return r
			}
		}
	}
	return proxy
}

func coverExtFromMime(mime string) string {
	switch {
	case strings.Contains(mime, "png"):
		return ".png"
	case strings.Contains(mime, "webp"):
		return ".webp"
	case strings.Contains(mime, "jpeg"), strings.Contains(mime, "jpg"):
		return ".jpg"
	case strings.Contains(mime, "gif"):
		return ".gif"
	default:
		return ""
	}
}
