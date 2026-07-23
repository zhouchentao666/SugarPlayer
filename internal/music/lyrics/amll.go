package lyrics

import (
	"errors"
	"net/url"
	"strings"

	"sugarplayer/internal/music/utils"
)

// amllDBBase is the public AMLL TTML lyrics database service used by
// CeruMusic (see src/renderer/src/store/GlobalPlayStatus.ts). It exposes
// per-platform, per-id Timed Text Markup Language (TTML) lyrics:
//   https://amll-ttml-db.stevexmh.net/[platform]/[musicID]
//
// The raw TTML is returned directly (not parsed): sugarplayer's frontend
// renders TTML natively via @applemusic-like-lyrics, which auto-detects the
// format and preserves word-level timings, translations and transliterations.
const amllDBBase = "https://amll-ttml-db.stevexmh.net"

// FetchAMLLLyric fetches the raw TTML lyrics for the given platform
// ("ncm" for NetEase, "qq" for QQ) and music id from the AMLL TTML database.
// On any failure (missing song, network error, empty/non-TTML result) it
// returns an error so callers can fall back to the platform's own lyric API.
func FetchAMLLLyric(platform, id string) (string, error) {
	if platform == "" || id == "" {
		return "", errors.New("amll: invalid args")
	}
	target := strings.Join([]string{amllDBBase, platform, url.QueryEscape(id)}, "/")
	body, err := utils.Get(target, utils.WithHeader("Referer", amllDBBase+"/"))
	if err != nil {
		return "", err
	}
	raw := string(body)
	// AMLL returns a 404 page / short "Not Found" text for missing songs; the
	// Get helper already errors on non-200, but guard against empty payloads
	// and obvious non-TTML responses (<tt ...> is the TTML root element).
	if len(strings.TrimSpace(raw)) < 100 || !strings.Contains(raw, "<tt") {
		return "", errors.New("amll: empty or invalid response")
	}
	return raw, nil
}
