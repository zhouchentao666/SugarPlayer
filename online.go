package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"sugarplayer/internal/music/core"
	"sugarplayer/internal/music/model"
)

// OnlineSong is the search result returned to the frontend.
type OnlineSong struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Artist    string `json:"artist"`
	Album     string `json:"album"`
	Cover     string `json:"cover"`
	Duration  int    `json:"duration"`
	Source    string `json:"source"`
	Extra     string `json:"extra"`
	Link      string `json:"link"`
	StreamURL string `json:"streamUrl"`
}

func toModelSong(s OnlineSong) *model.Song {
	ms := &model.Song{
		ID:       s.ID,
		Source:   s.Source,
		Name:     s.Name,
		Artist:   s.Artist,
		Album:    s.Album,
		Cover:    s.Cover,
		Duration: s.Duration,
		Link:     s.Link,
	}
	if s.Extra != "" {
		_ = json.Unmarshal([]byte(s.Extra), &ms.Extra)
	}
	return ms
}

func toOnlineSong(s model.Song, port int) OnlineSong {
	extra := ""
	if len(s.Extra) > 0 {
		if b, err := json.Marshal(s.Extra); err == nil {
			extra = string(b)
		}
	}
	stream := ""
	cover := s.Cover
	if s.ID != "" && s.Source != "" {
		q := url.Values{}
		q.Set("source", s.Source)
		q.Set("id", s.ID)
		if extra != "" {
			q.Set("extra", extra)
		}
		stream = fmt.Sprintf("http://127.0.0.1:%d/online?%s", port, q.Encode())
	}
	if cover != "" {
		cover = fmt.Sprintf("http://127.0.0.1:%d/cover?url=%s", port, url.QueryEscape(cover))
	}
	return OnlineSong{
		ID:        s.ID,
		Name:      s.Name,
		Artist:    s.Artist,
		Album:     s.Album,
		Cover:     cover,
		Duration:  s.Duration,
		Source:    s.Source,
		Extra:     extra,
		Link:      s.Link,
		StreamURL: stream,
	}
}

// OnlineSearch searches multiple music sources for the given keyword.
func (a *App) OnlineSearch(keyword string, sources []string) ([]OnlineSong, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return []OnlineSong{}, nil
	}

	available := make(map[string]bool)
	for _, s := range core.GetAllSourceNames() {
		available[s] = true
	}

	if len(sources) == 0 {
		sources = core.GetDefaultSourceNames()
	}
	// de-duplicate and keep only searchable sources
	seen := make(map[string]bool)
	resolved := make([]string, 0, len(sources))
	for _, s := range sources {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] || !available[s] {
			continue
		}
		if core.GetSearchFunc(s) == nil {
			continue
		}
		seen[s] = true
		resolved = append(resolved, s)
	}
	if len(resolved) == 0 {
		resolved = core.GetDefaultSourceNames()
	}

	var mu sync.Mutex
	var all []model.Song
	var wg sync.WaitGroup
	for _, src := range resolved {
		fn := core.GetSearchFunc(src)
		if fn == nil {
			continue
		}
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			res, err := fn(keyword)
			if err != nil {
				return
			}
			for i := range res {
				res[i].Source = s
			}
			mu.Lock()
			all = append(all, res...)
			mu.Unlock()
		}(src)
	}
	wg.Wait()

	out := make([]OnlineSong, 0, len(all))
	for _, s := range all {
		out = append(out, toOnlineSong(s, a.audio.port))
	}
	return out, nil
}

// OnlineQualityLevels 返回某在线歌曲在指定音源下可选的音质标识。
// 网易云 / QQ / 酷狗 / 酷我 通过 ZQ 网关支持 普通-无损-母带 三档。
func (a *App) OnlineQualityLevels(song OnlineSong) []string {
	switch song.Source {
	case "netease", "qq", "kugou", "kuwo":
		return core.GetQualityLevels(song.Source, toModelSong(song))
	default:
		return nil
	}
}

// OnlineLyric returns the LRC lyrics for an online song.
func (a *App) OnlineLyric(song OnlineSong) (string, error) {
	fn := core.GetLyricFunc(song.Source)
	if fn == nil {
		return "[00:00.00] 纯音乐 / 无歌词", nil
	}
	lrc, err := fn(toModelSong(song))
	if err != nil || strings.TrimSpace(lrc) == "" {
		return "[00:00.00] 纯音乐 / 无歌词", nil
	}
	return lrc, nil
}

// OnlineSources returns the available music sources with descriptions.
func (a *App) OnlineSources() []OnlineSource {
	names := core.GetAllSourceNames()
	excluded := map[string]bool{"local": true}
	sources := make([]OnlineSource, 0, len(names))
	for _, name := range names {
		if excluded[name] {
			continue
		}
		if core.GetSearchFunc(name) == nil {
			continue
		}
		enabled := !map[string]bool{"bilibili": true, "joox": true, "jamendo": true, "fivesing": true}[name]
		sources = append(sources, OnlineSource{
			ID:            name,
			Name:          core.GetSourceDescription(name),
			Enabled:       enabled,
			Recommend:     core.GetRecommendFunc(name) != nil,
			UserPlaylists: core.GetUserPlaylistsFunc(name) != nil,
		})
	}
	return sources
}

// OnlineSource describes a single music source for the UI.
type OnlineSource struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Enabled       bool   `json:"enabled"`
	Recommend     bool   `json:"recommend"`
	UserPlaylists bool   `json:"userPlaylists"`
}

// registerOnlineProxy adds a streaming proxy endpoint to the audio server so the
// WebView can play remote audio URLs through a local, CORS-friendly, range-aware proxy.
func (s *AudioServer) registerOnlineProxy() {
	s.mux.HandleFunc("/online", func(w http.ResponseWriter, r *http.Request) {
		source := strings.TrimSpace(r.URL.Query().Get("source"))
		id := strings.TrimSpace(r.URL.Query().Get("id"))
		extraRaw := strings.TrimSpace(r.URL.Query().Get("extra"))
		quality := strings.TrimSpace(r.URL.Query().Get("quality"))
		if source == "" || id == "" {
			http.NotFound(w, r)
			return
		}

		fn := core.GetDownloadFunc(source)
		if fn == nil {
			http.Error(w, "unsupported source", http.StatusBadRequest)
			return
		}

		song := &model.Song{ID: id, Source: source}
		if extraRaw != "" {
			_ = json.Unmarshal([]byte(extraRaw), &song.Extra)
		}
		// 用户选择的音质（仅 QQ/酷狗生效），失败由引擎回退到内置音质
		if quality != "" {
			if song.Extra == nil {
				song.Extra = map[string]string{}
			}
			song.Extra["quality"] = quality
		}

		downloadURL, err := fn(song)
		if err != nil || downloadURL == "" {
			http.Error(w, "failed to resolve audio url", http.StatusBadGateway)
			return
		}

		req, err := core.BuildSourceRequest("GET", downloadURL, source, r.Header.Get("Range"))
		if err != nil {
			http.Error(w, "failed to build request", http.StatusInternalServerError)
			return
		}
		for k, v := range r.Header {
			if strings.EqualFold(k, "Range") {
				continue
			}
			req.Header[k] = v
		}

		client := &http.Client{Timeout: 2 * time.Minute}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "upstream error", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		for k, v := range resp.Header {
			switch strings.ToLower(k) {
			case "transfer-encoding", "date", "connection", "content-length":
				continue
			default:
				w.Header()[k] = v
			}
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if w.Header().Get("Accept-Ranges") == "" {
			w.Header().Set("Accept-Ranges", "bytes")
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	})
}

// registerCoverProxy proxies remote cover images through a local endpoint so the
// WebView can display them without running into cross-origin/referer restrictions.
func (s *AudioServer) registerCoverProxy() {
	s.mux.HandleFunc("/cover", func(w http.ResponseWriter, r *http.Request) {
		u := strings.TrimSpace(r.URL.Query().Get("url"))
		if u == "" || !strings.HasPrefix(u, "http") {
			http.NotFound(w, r)
			return
		}

		req, err := http.NewRequest("GET", u, nil)
		if err != nil {
			http.Error(w, "bad url", http.StatusBadRequest)
			return
		}
		req.Header.Set("User-Agent", core.UA_Common)
		if referer := coverReferer(u); referer != "" {
			req.Header.Set("Referer", referer)
		}

		client := &http.Client{Timeout: 20 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "fetch failed", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			http.Error(w, "upstream error", http.StatusBadGateway)
			return
		}

		ct := strings.TrimSpace(resp.Header.Get("Content-Type"))
		if strings.Contains(ct, "text/html") || strings.Contains(ct, "text/plain") {
			http.Error(w, "not an image", http.StatusBadGateway)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Cache-Control", "public, max-age=21600")
		if ct != "" {
			w.Header().Set("Content-Type", ct)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.Copy(w, resp.Body)
	})
}

func coverReferer(u string) string {
	switch {
	case strings.Contains(u, "163.com"):
		return core.Ref_Netease
	case strings.Contains(u, "qq.com"):
		return "http://y.qq.com"
	case strings.Contains(u, "migu.cn"):
		return core.Ref_Migu
	default:
		return ""
	}
}

// SetPlatformCookies stores music-platform login cookies (the go-music-dl
// "Cookies" feature) into the shared core cookie manager so that subsequent
// searches / downloads / playback use the logged-in state. Persistence is
// handled by the frontend via SaveConfig (config.json); this only updates the
// runtime cookie manager consumed by core.GetSearchFunc / GetDownloadFunc.
func (a *App) SetPlatformCookies(cookies map[string]string) error {
	core.CM.SetAll(cookies)
	return nil
}

// GetPlatformCookies returns the currently configured platform cookies.
func (a *App) GetPlatformCookies() (map[string]string, error) {
	return core.CM.GetAll(), nil
}

// SwitchSongSource tries to find a playable source for the given online song.
// If the current source is already playable it returns the song unchanged;
// otherwise it searches other sources for a matching song and returns the first
// playable alternative (or the original song if none is found). This backs the
// "自动选择无效音源并批量换源" feature.
func (a *App) SwitchSongSource(song OnlineSong) (OnlineSong, error) {
	model := toModelSong(song)
	if core.ValidatePlayable(model) {
		return song, nil
	}
	keyword := strings.TrimSpace(song.Name + " " + song.Artist)
	if keyword == "" {
		return song, nil
	}
	for _, src := range core.GetAllSourceNames() {
		if src == song.Source || src == "local" {
			continue
		}
		fn := core.GetSearchFunc(src)
		if fn == nil {
			continue
		}
		res, err := fn(keyword)
		if err != nil {
			continue
		}
		for i := range res {
			if core.CalcSongSimilarity(song.Name, song.Artist, res[i].Name, res[i].Artist) < 0.6 {
				continue
			}
			if core.ValidatePlayable(&res[i]) {
				return toOnlineSong(res[i], a.audio.port), nil
			}
		}
	}
	return song, nil
}
