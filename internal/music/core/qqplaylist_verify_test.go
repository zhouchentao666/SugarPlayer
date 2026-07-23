package core

import (
	"fmt"
	"testing"

	"sugarplayer/internal/music/model"
	"sugarplayer/internal/music/qq"
)

func TestQQPlaylistSongResolve(t *testing.T) {
	mids := []string{"000xdZuV2LcQ19", "003U0dF50ZL8Y7", "002h62C40NbkWI"}
	for _, m := range mids {
		song := &model.Song{ID: m, Source: "qq", Extra: map[string]string{"songmid": m}}
		// direct Go engine (the fallback path)
		u, err := qq.New("").GetDownloadURL(song)
		fmt.Printf("GOENGINE mid=%s url=%q err=%v\n", m, u, err)
		// full qzOrFallback path
		fn := GetDownloadFunc("qq")
		u2, err2 := fn(&model.Song{ID: m, Source: "qq"})
		fmt.Printf("FULLPATH mid=%s url=%q err=%v\n", m, u2, err2)
	}
}
