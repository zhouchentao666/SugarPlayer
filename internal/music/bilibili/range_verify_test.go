package bilibili

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestBilibiliRangeBehavior(t *testing.T) {
	b := New("")
	// use a well-known BV with a short audio track
	songs, err := b.Search("晚安")
	if err != nil {
		t.Skipf("search failed: %v", err)
	}
	if len(songs) == 0 {
		t.Skip("no bilibili songs found")
	}
	s := songs[0]
	u, err := b.GetDownloadURL(&s)
	if err != nil || u == "" {
		t.Skipf("no download url: %v", err)
	}
	fmt.Printf("URL=%s\n", u)

	client := &http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("Range", "bytes=0-1023")
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Referer", Referer)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("range request failed: %v", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	fmt.Printf("STATUS=%d Accept-Ranges=%q Content-Range=%q Content-Length=%q\n",
		resp.StatusCode, resp.Header.Get("Accept-Ranges"), resp.Header.Get("Content-Range"), resp.Header.Get("Content-Length"))
}
