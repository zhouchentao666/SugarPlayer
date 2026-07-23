package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"

	"sugarplayer/internal/music/core"
)

func kgSig(params string) string {
	parts := strings.Split(params, "&")
	sort.Strings(parts)
	joined := strings.Join(parts, "")
	sum := md5.Sum([]byte("OIlwieks28dk2k092lski2UIkp" + joined + "OIlwieks28dk2k092lski2UIkp"))
	return hex.EncodeToString(sum[:])
}

func TestKgCommentReal(t *testing.T) {
	fn := core.GetSearchFunc("kugou")
	if fn == nil {
		t.Fatal("no kugou search func")
	}
	songs, err := fn("周杰伦")
	if err != nil {
		t.Fatal(err)
	}
	if len(songs) == 0 {
		t.Fatal("no kugou songs")
	}
	hash := songs[0].ID
	t.Logf("kugou hash=%s name=%s", hash, songs[0].Name)

	ts := time.Now().UnixMilli()
	params := fmt.Sprintf(
		"dfid=0&mid=16249512204336365674023395779019&clienttime=%d&uuid=0&extdata=%s&appid=1005&code=fc4be23b4e972707f36b8a828a93ba8a&schash=%s&clientver=11409&p=%d&clienttoken=&pagesize=%d&ver=10&kugouid=0",
		ts, hash, hash, 1, 20,
	)
	endpoint := fmt.Sprintf("http://m.comment.service.kugou.com/r/v1/rank/newest?%s&signature=%s", params, kgSig(params))
	req, _ := http.NewRequest(http.MethodGet, endpoint, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36 Edg/107.0.1418.24")
	client := &http.Client{Timeout: 25 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	t.Logf("status=%d body=%s", resp.StatusCode, string(body))
	_ = url.Values{}
}
