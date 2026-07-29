// Package qzgw implements playback-URL resolution through the closed-source
// QZ gateway (api.qz.shiqianjiang.cn). The gateway endpoint and its access key
// are NOT open source, so this entire package is excluded from version control
// (see .gitignore: internal/music/qzgw/). It is physically present in local
// builds only.
//
// Usage:
//   - Call qzgw.SetKey(userProvidedKey) to activate the gateway.
//   - Call qzgw.IsUnlocked() to check whether a valid key has been supplied.
//   - Use qzgw.QzOrFallback(source, fallback) to wrap a download function.
package qzgw

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"sugarplayer/internal/music/model"
)

// validKey is the software-unlock password that the end user must enter in the
// online settings before the QZ gateway is used. It is stored exclusively in
// this closed-source package and only gates whether the gateway is active.
const validKey = "7sK9pR2vG5bQ8nD3zX6cT1mF4jH0aLw"

// qzGatewayKey is the credential sent to the QZ gateway endpoint itself. The
// gateway accepts this fixed key (the same one the previous open-source client
// used), so once the user has unlocked the gateway with validKey we always
// authenticate the request with qzGatewayKey regardless of the entered password.
const qzGatewayKey = "testkey"

// qzUA mirrors core.UA_Common; kept local so this package has no dependency on
// the open-source core package internals.
const qzUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36"

// qzGatewayBase is the playback-resolution endpoint.
const qzGatewayBase = "https://api.qz.shiqianjiang.cn/music/url"

// qzGatewaySource maps SugarPlayer's internal source names to the gateway's
// "source" query parameter.
var qzGatewaySource = map[string]string{
	"netrease": "wy",
	"qq":      "tx",
	"kugou":   "kg",
	"kuwo":    "kw",
}

var (
	mu         sync.RWMutex
	currentKey string
)

// SetKey validates and stores the QZ gateway key. Returns true if the key
// matches the required secret.
func SetKey(key string) bool {
	mu.Lock()
	defer mu.Unlock()
	if key == validKey {
		currentKey = key
		return true
	}
	currentKey = ""
	return false
}

// IsUnlocked reports whether a valid QZ gateway key has been supplied.
func IsUnlocked() bool {
	mu.RLock()
	defer mu.RUnlock()
	return currentKey == validKey
}

// zSupported reports whether a source is served by the QZ gateway.
func qzSupported(source string) bool {
	_, ok := qzGatewaySource[source]
	return ok
}

// qzGatewayQuality normalizes a user-facing quality id to one accepted by the
// gateway. The three user tiers are 普通(standard) / 无损(lossless) / 母带(hires).
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
// returns the playable URL. The stored key is used for authentication via
// either the query parameter (tx/QQ) or the X-API-KEY header (all others).
func resolveQZOnce(source, quality, songID string) (string, error) {
	if !IsUnlocked() {
		return "", errors.New("qz gateway: key not unlocked")
	}

	gw, ok := qzGatewaySource[source]
	if !ok {
		return "", fmt.Errorf("qz gateway: unsupported source %q", source)
	}

	q := url.Values{}
	q.Set("songId", songID)
	q.Set("quality", quality)
	q.Set("source", gw)
	if gw == "tx" {
		q.Set("key", qzGatewayKey)
	}
	endpoint := qzGatewayBase + "?" + q.Encode()

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	if gw != "tx" {
		req.Header.Set("X-API-KEY", qzGatewayKey)
	}
	req.Header.Set("User-Agent", qzUA)

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
// resolution.
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
// QZ gateway. song.Extra["quality"] selects the tier; on failure it retries
// other tiers before giving up.
func ResolveQZDownloadURL(source string, song *model.Song) (string, error) {
	if !IsUnlocked() {
		return "", errors.New("qz gateway: key not unlocked")
	}
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

// QzOrFallback returns a download func that prefers the QZ gateway but falls
// back to the original provider implementation on any failure, so playback
// degrades gracefully if the gateway is unreachable or the key is not set.
func QzOrFallback(source string, fallback func(*model.Song) (string, error)) func(*model.Song) (string, error) {
	return func(s *model.Song) (string, error) {
		if u, err := ResolveQZDownloadURL(source, s); err == nil && u != "" {
			return u, nil
		}
		return fallback(s)
	}
}
