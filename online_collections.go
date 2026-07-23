package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"sugarplayer/internal/music/core"
	"sugarplayer/internal/music/model"
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

// OnlineCategoryItem is a single selectable playlist category (e.g. 华语 / 流行).
type OnlineCategoryItem struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Hot    bool   `json:"hot"`
	Source string `json:"source"`
}

// OnlineCategoryGroup groups categories under a heading (e.g. 语种 / 风格).
type OnlineCategoryGroup struct {
	Name       string               `json:"name"`
	Categories []OnlineCategoryItem `json:"categories"`
}

// OnlineCategorySource holds the full playlist-category tree for one music source.
type OnlineCategorySource struct {
	Source string                `json:"source"`
	Name   string                `json:"name"`
	Groups []OnlineCategoryGroup `json:"groups"`
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

// OnlinePlaylistCategories returns the full playlist-category tree aggregated
// from every source that supports categories (netease / qq / kugou / kuwo / ...).
// When sources is empty, all category-capable sources are used.
func (a *App) OnlinePlaylistCategories(sources []string) ([]OnlineCategorySource, error) {
	if len(sources) == 0 {
		sources = core.GetPlaylistCategorySourceNames()
	}
	var mu sync.Mutex
	views := make([]OnlineCategorySource, 0, len(sources))
	var wg sync.WaitGroup
	for _, src := range sources {
		fn := core.GetPlaylistCategoriesFunc(src)
		if fn == nil {
			continue
		}
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			cats, err := fn()
			if err != nil || len(cats) == 0 {
				return
			}
			view := buildOnlineCategorySource(s, cats)
			mu.Lock()
			views = append(views, view)
			mu.Unlock()
		}(src)
	}
	wg.Wait()
	return views, nil
}

func buildOnlineCategorySource(source string, cats []model.PlaylistCategory) OnlineCategorySource {
	groupIndex := make(map[string]int)
	groups := make([]OnlineCategoryGroup, 0)
	for _, c := range cats {
		name := strings.TrimSpace(c.Name)
		if name == "" {
			continue
		}
		groupName := strings.TrimSpace(c.Group)
		if groupName == "" {
			groupName = "其他"
		}
		idx, ok := groupIndex[groupName]
		if !ok {
			idx = len(groups)
			groupIndex[groupName] = idx
			groups = append(groups, OnlineCategoryGroup{Name: groupName})
		}
		groups[idx].Categories = append(groups[idx].Categories, OnlineCategoryItem{
			ID:     strings.TrimSpace(c.ID),
			Name:   name,
			Hot:    c.Hot,
			Source: source,
		})
	}
	return OnlineCategorySource{
		Source: source,
		Name:   core.GetSourceDescription(source),
		Groups: groups,
	}
}

// OnlineCategoryPlaylists returns the playlists for a given source + category.
// categoryID / categoryName come from an OnlineCategoryItem; an empty category
// falls back to "全部".
func (a *App) OnlineCategoryPlaylists(source, categoryID, categoryName string) ([]OnlineCollection, error) {
	source = strings.TrimSpace(source)
	categoryID = strings.TrimSpace(categoryID)
	categoryName = strings.TrimSpace(categoryName)
	if categoryName == "" {
		categoryName = categoryID
	}
	if categoryName == "" {
		categoryName = "全部"
	}
	fn := core.GetCategoryPlaylistsFunc(source)
	if source == "" || fn == nil {
		return nil, fmt.Errorf("该音源不支持歌单分类")
	}
	pls, err := fn(categoryID, 1, 120)
	if err != nil {
		return nil, err
	}
	for i := range pls {
		pls[i].Source = source
	}
	out := make([]OnlineCollection, 0, len(pls))
	for _, p := range pls {
		out = append(out, toOnlineCollection(p, "playlist", a.audio.port))
	}
	return out, nil
}
