package core

import (
	"fmt"
	"testing"

	"sugarplayer/internal/music/model"
)

func TestResolveQZKugouVerify(t *testing.T) {
	vip := &model.Song{
		Source: "kugou",
		ID:     "B3A52A7A958BF0AED0EBFBA2E9A818B7", // 晴天 (privilege=10)
		Extra:  map[string]string{"quality": "standard"},
	}
	free := &model.Song{
		Source: "kugou",
		ID:     "5126B551BD6C3E4C82E64474E223C94B", // free
		Extra:  map[string]string{"quality": "standard"},
	}
	for name, s := range map[string]*model.Song{"vip-standard": vip, "free-standard": free} {
		u, err := ResolveQZDownloadURL("kugou", s)
		fmt.Printf("[%s] url=%q err=%v\n", name, u, err)
		if err != nil {
			t.Logf("%s: error (will rely on Go fallback): %v", name, err)
		}
	}
}
