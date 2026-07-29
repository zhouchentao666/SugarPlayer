package main

import (
	"os"
	"path/filepath"
	"strings"

	"sugarplayer/internal/music/qzgw"
)

// qzKeyPath returns the path of the file that stores the user-provided QZ
// gateway key. The QZ gateway is closed source and requires a valid key before
// it will resolve any playback URL; we persist the key locally so it survives
// restarts without re-prompting the user.
func qzKeyPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	return filepath.Join(configDir, "SugarMusic", "qz_key")
}

// loadQZKey restores the persisted QZ gateway key (if any) into the closed-source
// qzgw package so QZ playback works immediately after a restart.
func (a *App) loadQZKey() {
	data, err := os.ReadFile(qzKeyPath())
	if err != nil {
		qzgw.SetKey("")
		return
	}
	qzgw.SetKey(strings.TrimSpace(string(data)))
}

// SetQZKey validates the user-provided QZ gateway key. On success the key is
// persisted locally and the QZ gateway is unlocked for playback / download; on
// failure the gateway stays locked and any previously stored key is cleared.
func (a *App) SetQZKey(key string) bool {
	ok := qzgw.SetKey(strings.TrimSpace(key))
	if ok {
		path := qzKeyPath()
		if mkErr := os.MkdirAll(filepath.Dir(path), 0755); mkErr == nil {
			_ = os.WriteFile(path, []byte(strings.TrimSpace(key)), 0644)
		}
	} else {
		// 验证失败：清除本地存储的密钥，确保网关保持锁定
		_ = os.Remove(qzKeyPath())
	}
	return ok
}

// GetQZKey returns the currently stored QZ gateway key (empty if none). It is
// used by the UI to prefill the input field.
func (a *App) GetQZKey() string {
	data, err := os.ReadFile(qzKeyPath())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// QZIsUnlocked reports whether the QZ gateway is active (a valid key was set).
func (a *App) QZIsUnlocked() bool {
	return qzgw.IsUnlocked()
}
