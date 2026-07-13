package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/guohuiyuan/go-music-dl/core"
	"github.com/guohuiyuan/music-lib/model"
)

// OnlineCollection is a playlist or album returned by the online music sources.
// It is the unit that can be searched, recommended, listed as "my playlists",
// opened to show its songs, and pinned to the sidebar.
type OnlineCollection struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Cover      string `json:"cover"`
	Source     string `json:"source"`
	Link       string `json:"link"`
	Kind       string `json:"kind"` // "playlist" | "album"
	Creator    string `json:"creator"`
	TrackCount int    `json:"trackCount"`
	Extra      string `json:"extra"`
}

func toOnlineCollection(p model.Playlist, kind string, port int) OnlineCollection {
	cover := p.Cover
	if cover != "" {
		cover = fmt.Sprintf("http://127.0.0.1:%d/cover?url=%s", port, url.QueryEscape(cover))
	}
	extra := ""
	if len(p.Extra) > 0 {
		if b, err := json.Marshal(p.Extra); err == nil {
			extra = string(b)
		}
	}
	return OnlineCollection{
		ID:         p.ID,
		Name:       p.Name,
		Cover:      cover,
		Source:     p.Source,
		Link:       p.Link,
		Kind:       kind,
		Creator:    p.Creator,
		TrackCount: p.TrackCount,
		Extra:      extra,
	}
}

// OnlineRecommendPlaylists returns daily recommended playlists aggregated from
// the sources that support recommendations (netease / qq / kugou / kuwo).
func (a *App) OnlineRecommendPlaylists(sources []string) ([]OnlineCollection, error) {
	if len(sources) == 0 {
		sources = core.GetRecommendSourceNames()
	}
	var mu sync.Mutex
	var all []model.Playlist
	var wg sync.WaitGroup
	for _, src := range sources {
		fn := core.GetRecommendFunc(src)
		if fn == nil {
			continue
		}
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			res, err := fn()
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
	out := make([]OnlineCollection, 0, len(all))
	for _, p := range all {
		out = append(out, toOnlineCollection(p, "playlist", a.audio.port))
	}
	return out, nil
}

// OnlineUserPlaylists returns the logged-in user's playlists (created / collected
// / liked) from the sources that support it (netease / qq / kugou / soda). A valid
// platform cookie must be configured via the online settings for results to appear.
func (a *App) OnlineUserPlaylists(sources []string) ([]OnlineCollection, error) {
	if len(sources) == 0 {
		sources = core.GetUserPlaylistSourceNames()
	}
	var mu sync.Mutex
	var all []model.Playlist
	var wg sync.WaitGroup
	for _, src := range sources {
		fn := core.GetUserPlaylistsFunc(src)
		if fn == nil {
			continue
		}
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			res, err := fn(0, 200)
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
	out := make([]OnlineCollection, 0, len(all))
	for _, p := range all {
		out = append(out, toOnlineCollection(p, "playlist", a.audio.port))
	}
	return out, nil
}

// OnlineSearchCollections searches playlists or albums (kind = "playlist" | "album")
// across the selected sources and returns a unified list of collections.
func (a *App) OnlineSearchCollections(keyword string, kind string, sources []string) ([]OnlineCollection, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return []OnlineCollection{}, nil
	}
	isAlbum := kind == "album"
	var getter func(string) core.SearchPlaylistFunc
	var names []string
	if isAlbum {
		getter = core.GetAlbumSearchFunc
		names = core.GetAlbumSourceNames()
	} else {
		getter = core.GetPlaylistSearchFunc
		names = core.GetPlaylistSourceNames()
	}

	available := make(map[string]bool)
	for _, s := range names {
		available[s] = true
	}
	if len(sources) == 0 {
		sources = names
	}
	seen := make(map[string]bool)
	resolved := make([]string, 0, len(sources))
	for _, s := range sources {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] || !available[s] {
			continue
		}
		if getter(s) == nil {
			continue
		}
		seen[s] = true
		resolved = append(resolved, s)
	}
	if len(resolved) == 0 {
		resolved = names
	}

	var mu sync.Mutex
	var all []model.Playlist
	var wg sync.WaitGroup
	for _, src := range resolved {
		fn := getter(src)
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

	k := "playlist"
	if isAlbum {
		k = "album"
	}
	out := make([]OnlineCollection, 0, len(all))
	for _, p := range all {
		out = append(out, toOnlineCollection(p, k, a.audio.port))
	}
	return out, nil
}

// OnlineCollectionSongs returns the songs contained in a playlist or album.
func (a *App) OnlineCollectionSongs(col OnlineCollection) ([]OnlineSong, error) {
	var fn func(string) ([]model.Song, error)
	if col.Kind == "album" {
		fn = core.GetAlbumDetailFunc(col.Source)
	} else {
		fn = core.GetPlaylistDetailFunc(col.Source)
	}
	if fn == nil {
		return nil, fmt.Errorf("该音源不支持%s歌曲获取", map[bool]string{true: "专辑", false: "歌单"}[col.Kind == "album"])
	}
	songs, err := fn(col.ID)
	if err != nil {
		return nil, err
	}
	out := make([]OnlineSong, 0, len(songs))
	for _, s := range songs {
		out = append(out, toOnlineSong(s, a.audio.port))
	}
	return out, nil
}
