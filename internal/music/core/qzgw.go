package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"sugarplayer/internal/music/model"
)

// qzGatewayBase is the playback-resolution endpoint used by the ZQ plugins
// shipped in 新建文件夹/zq_*.js. Those plugins are thin clients that proxy
// audio-URL resolution to this gateway; calling it directly from Go lets us
// replace the built-in netease/qq/kugou/kuwo playback (with quality switching
// and "download uses current quality") without changing the rest of the
// online-search / stream / download pipeline.
const qzGatewayBase = "https://api.qz.shiqianjiang.cn/music/url"

// qzGatewaySource maps SugarPlayer's internal source names to the gateway's
// "source" query parameter.
var qzGatewaySource = map[string]string{
	"netease": "wy",
	"qq":      "tx",
	"kugou":   "kg",
	"kuwo":    "kw",
}

// qzSupported reports whether a source is served by the QZ gateway.
func qzSupported(source string) bool {
	_, ok := qzGatewaySource[source]
	return ok
}

// qzGatewayQuality normalizes a user-facing quality id to one accepted by the
// gateway. The three user tiers are 普通(standard) / 无损(lossless) / 母带(hires).
//
// NOTE: the gateway's per-source semantics differ:
//   - netease / qq / kuwo accept "standard" / "lossless" / "hires"
//   - kugou only accepts "flac" (not "lossless") and "hires" (and "128"/"320");
//     "standard"/"lossless" are rejected with 解析失败. So for kugou we map the
//     无损 tier to "flac".
func qzGatewayQuality(source, q string) string {
	switch strings.ToLower(strings.TrimSpace(q)) {
	case "standard", "normal", "128", "low", "":
		return "standard"
	case "exhigh", "high", "hq", "320":
		return "exhigh"
	case "lossless", "flac", "sq":
		if source == "kugou" {
			return "flac"
		}
		return "lossless"
	case "hires", "hi-res", "hr":
		return "hires"
	case "jymaster", "master", "jyeffect", "sky":
		return strings.ToLower(strings.TrimSpace(q))
	default:
		return "standard"
	}
}

type qzGatewayResp struct {
	URL  string `json:"url"`
	Data struct {
		URL string `json:"url"`
	} `json:"data"`
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	Detail string `json:"detail"`
}

// resolveQZOnce performs a single gateway request for the given quality and
// returns the playable URL. It returns an error when the gateway is
// unreachable, returns a non-200 status, or does not yield an http(s) URL.
func resolveQZOnce(source, quality, songID string) (string, error) {
	gw, ok := qzGatewaySource[source]
	if !ok {
		return "", fmt.Errorf("qz gateway: unsupported source %q", source)
	}

	q := url.Values{}
	q.Set("songId", songID)
	q.Set("quality", quality)
	q.Set("source", gw)
	if gw == "tx" {
		q.Set("key", "testkey")
	}
	endpoint := qzGatewayBase + "?" + q.Encode()

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	// 与对应插件保持一致：
	//   qq(tx)  用 query 参数 key=testkey
	//   wy/kg/kw 用请求头 X-API-KEY: testkey
	// 缺了 X-API-KEY 会导致网关返回「当前密钥权限不足 / 解析失败」，酷狗/酷我无法播放。
	if gw != "tx" {
		req.Header.Set("X-API-KEY", "testkey")
	}
	req.Header.Set("User-Agent", UA_Common)

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("qz gateway: http %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var r qzGatewayResp
	if err := json.Unmarshal(body, &r); err != nil {
		return "", fmt.Errorf("qz gateway: bad response: %w (body=%s)", err, truncate(string(body), 200))
	}
	out := r.URL
	if out == "" {
		out = r.Data.URL
	}
	if out == "" || !strings.HasPrefix(out, "http") {
		msg := r.Msg
		if msg == "" {
			msg = r.Detail
		}
		if msg == "" {
			msg = "no url returned"
		}
		return "", fmt.Errorf("qz gateway: %s", msg)
	}
	return out, nil
}

// qzRetryQualities lists the qualities to try, in order, for one gateway
// resolution. We always start with the user-requested quality, then fall back
// to whatever the gateway can actually serve for that source (so a VIP kugou
// song requested at 普通 still plays at 母带/无损 instead of failing).
func qzRetryQualities(source, requested string) []string {
	order := []string{requested, "hires", "flac", "lossless", "exhigh", "standard", "320", "128"}
	seen := make(map[string]struct{}, len(order))
	result := make([]string, 0, len(order))
	for _, q := range order {
		if q == "" {
			continue
		}
		if _, ok := seen[q]; ok {
			continue
		}
		seen[q] = struct{}{}
		result = append(result, q)
	}
	return result
}

// ResolveQZDownloadURL resolves the playable audio URL for a song through the
// QZ gateway. song.Extra["quality"] selects the tier (standard/lossless/hires);
// on failure it automatically retries other tiers before giving up so the
// caller's fallback (the original provider) can take over.
func ResolveQZDownloadURL(source string, song *model.Song) (string, error) {
	requested := qzGatewayQuality(source, songExtraQuality(song))
	var lastErr error
	for _, q := range qzRetryQualities(source, requested) {
		u, err := resolveQZOnce(source, q, song.ID)
		if err == nil && u != "" {
			return u, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = errors.New("qz gateway: no url returned")
	}
	return "", lastErr
}

func songExtraQuality(song *model.Song) string {
	if song == nil || song.Extra == nil {
		return ""
	}
	return song.Extra["quality"]
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// qzOrFallback returns a download func that prefers the QZ gateway but falls
// back to the original provider implementation on any failure, so playback
// degrades gracefully if the gateway is unreachable.
func qzOrFallback(source string, fallback func(*model.Song) (string, error)) func(*model.Song) (string, error) {
	return func(s *model.Song) (string, error) {
		if u, err := ResolveQZDownloadURL(source, s); err == nil && u != "" {
			return u, nil
		}
		return fallback(s)
	}
}
