package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"sugarplayer/internal/music/core"
)

// ---------- online music cache (runtime configuration) ----------
// These mirror the persisted ConfigSettings fields and are initialised from the
// saved config at startup, then updated live by the settings UI.
var (
	onlineCacheMu      sync.RWMutex
	onlineCacheEnabled bool  = true
	onlineCacheMax     int64 = 2048 * 1024 * 1024 // bytes; 0 means unlimited
)

func onlineCacheEnabledNow() bool {
	onlineCacheMu.RLock()
	defer onlineCacheMu.RUnlock()
	return onlineCacheEnabled
}

func onlineCacheMaxNow() int64 {
	onlineCacheMu.RLock()
	defer onlineCacheMu.RUnlock()
	return onlineCacheMax
}

func setOnlineCacheEnabledFlag(v bool) {
	onlineCacheMu.Lock()
	onlineCacheEnabled = v
	onlineCacheMu.Unlock()
}

func setOnlineCacheMaxMB(maxMB int) {
	onlineCacheMu.Lock()
	if maxMB < 0 {
		maxMB = 0
	}
	if maxMB == 0 {
		onlineCacheMax = 0
	} else {
		onlineCacheMax = int64(maxMB) * 1024 * 1024
	}
	onlineCacheMu.Unlock()
}

func onlineCacheDir() string {
	cd, err := os.UserCacheDir()
	if err != nil {
		cd = os.TempDir()
	}
	return filepath.Join(cd, "SugarMusic", "cache", "online")
}

// onlineCacheKey derives a stable file name from the request parameters. The
// quality is included because it changes the resolved audio URL.
func onlineCacheKey(source, id, quality, extra string) string {
	sum := sha256.Sum256([]byte(source + "\x00" + id + "\x00" + quality + "\x00" + extra))
	return hex.EncodeToString(sum[:])
}

// onlineCacheFiles lists the cached .cache files and their total size.
func onlineCacheFiles(dir string) ([]string, int64, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, 0, err
	}
	var paths []string
	var total int64
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".cache") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		paths = append(paths, filepath.Join(dir, e.Name()))
		total += info.Size()
	}
	return paths, total, nil
}

// serveOnlineCache attempts to serve a previously cached file. It returns true
// when the response was written (Range-aware via http.ServeContent).
func serveOnlineCache(w http.ResponseWriter, r *http.Request, key string) bool {
	if !onlineCacheEnabledNow() {
		return false
	}
	dir := onlineCacheDir()
	cachePath := filepath.Join(dir, key+".cache")
	info, err := os.Stat(cachePath)
	if err != nil {
		return false
	}
	f, err := os.Open(cachePath)
	if err != nil {
		return false
	}
	defer f.Close()

	ctype := "audio/mpeg"
	if b, err := os.ReadFile(filepath.Join(dir, key+".meta")); err == nil && len(strings.TrimSpace(string(b))) > 0 {
		ctype = strings.TrimSpace(string(b))
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Type", ctype)
	w.Header().Set("X-Cache", "HIT")
	http.ServeContent(w, r, key, info.ModTime(), f)
	return true
}

var (
	onlineCacheInflight   = make(map[string]bool)
	onlineCacheInflightMu sync.Mutex
)

func onlineCacheMarkInflight(key string) bool {
	onlineCacheInflightMu.Lock()
	defer onlineCacheInflightMu.Unlock()
	if onlineCacheInflight[key] {
		return true
	}
	onlineCacheInflight[key] = true
	return false
}

func onlineCacheClearInflight(key string) {
	onlineCacheInflightMu.Lock()
	delete(onlineCacheInflight, key)
	onlineCacheInflightMu.Unlock()
}

// prefetchOnlineCache downloads the full audio into the cache in the background
// so that subsequent plays are served locally.
func prefetchOnlineCache(key, source, downloadURL string) {
	if !onlineCacheEnabledNow() {
		return
	}
	if onlineCacheMarkInflight(key) {
		return
	}
	defer onlineCacheClearInflight(key)

	dir := onlineCacheDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}
	cachePath := filepath.Join(dir, key+".cache")
	if _, err := os.Stat(cachePath); err == nil {
		return // already cached
	}
	tmpPath := filepath.Join(dir, key+".tmp")

	req, err := core.BuildSourceRequest("GET", downloadURL, source, "")
	if err != nil {
		return
	}
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return
	}
	ctype := resp.Header.Get("Content-Type")

	out, err := os.Create(tmpPath)
	if err != nil {
		return
	}
	n, err := io.Copy(out, resp.Body)
	out.Close()
	if err != nil || n == 0 {
		os.Remove(tmpPath)
		return
	}
	_ = os.WriteFile(filepath.Join(dir, key+".meta"), []byte(ctype), 0644)
	if err := os.Rename(tmpPath, cachePath); err != nil {
		os.Remove(tmpPath)
		return
	}
	enforceOnlineCacheMax()
}

// enforceOnlineCacheMax evicts least-recently-used cache files until the total
// size is within the configured limit.
func enforceOnlineCacheMax() {
	max := onlineCacheMaxNow()
	if max <= 0 {
		return
	}
	dir := onlineCacheDir()
	paths, total, err := onlineCacheFiles(dir)
	if err != nil || total <= max {
		return
	}
	type fi struct {
		path string
		size int64
		mod  time.Time
	}
	files := make([]fi, 0, len(paths))
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		files = append(files, fi{p, info.Size(), info.ModTime()})
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].mod.Before(files[j].mod)
	})
	for _, f := range files {
		if total <= max {
			break
		}
		os.Remove(f.path)
		os.Remove(strings.TrimSuffix(f.path, ".cache") + ".meta")
		total -= f.size
	}
}

// ---------- App bindings for the settings UI ----------

// GetOnlineCacheSize returns the current online music cache size in bytes.
func (a *App) GetOnlineCacheSize() int64 {
	_, total, err := onlineCacheFiles(onlineCacheDir())
	if err != nil {
		return 0
	}
	return total
}

// ClearOnlineCache removes all cached online music files.
func (a *App) ClearOnlineCache() error {
	dir := onlineCacheDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".cache") || strings.HasSuffix(name, ".meta") || strings.HasSuffix(name, ".tmp") {
			_ = os.Remove(filepath.Join(dir, name))
		}
	}
	return nil
}

// SetOnlineCacheEnabled toggles the online music cache at runtime. Persistence
// is handled by the frontend via SaveConfig.
func (a *App) SetOnlineCacheEnabled(enabled bool) {
	setOnlineCacheEnabledFlag(enabled)
}

// SetOnlineCacheMaxSize sets the cache cap in megabytes (0 = unlimited).
func (a *App) SetOnlineCacheMaxSize(maxMB int) {
	setOnlineCacheMaxMB(maxMB)
}

// loadOnlineCacheConfig initialises the runtime cache flags from the persisted
// config so that the limits configured in a previous session are honoured.
func (a *App) loadOnlineCacheConfig() {
	cfg, err := a.LoadConfig()
	if err != nil {
		return
	}
	setOnlineCacheEnabledFlag(cfg.Settings.OnlineCacheEnabled)
	setOnlineCacheMaxMB(cfg.Settings.OnlineCacheMaxSizeMB)
}
