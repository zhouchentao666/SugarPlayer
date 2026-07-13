package main

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/guohuiyuan/go-music-dl/core"
	"github.com/guohuiyuan/music-lib/model"
	qrcode "github.com/skip2/go-qrcode"
)

// QRLoginSources returns the music platforms that support QR-code login.
// This includes WeChat login for QQ Music via the special "qq_wx" source.
func (a *App) QRLoginSources() []string {
	return core.GetQRLoginSourceNames()
}

// CreateQRLogin creates a QR login session for the given platform. When the
// upstream session does not carry a ready-made image, a PNG data URL is
// generated locally from the QR content URL so the frontend can render it
// directly without any extra dependency.
func (a *App) CreateQRLogin(source string) (model.QRLoginSession, error) {
	fn := core.GetQRLoginCreateFunc(source)
	if fn == nil {
		return model.QRLoginSession{}, fmt.Errorf("unsupported qr login source: %s", source)
	}
	session, err := fn()
	if err != nil {
		return model.QRLoginSession{}, err
	}
	if session == nil {
		return model.QRLoginSession{}, fmt.Errorf("empty qr login session")
	}
	if strings.TrimSpace(session.ImageURL) == "" && strings.TrimSpace(session.URL) != "" {
		if png, encErr := qrcode.Encode(session.URL, qrcode.Medium, 256); encErr == nil {
			session.ImageURL = "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
		}
	}
	return *session, nil
}

// CheckQRLogin polls the login status for a previously created QR session.
// On success it extracts the login cookie and injects it into the shared core
// cookie manager so subsequent searches / playback use the logged-in state; the
// resolved cookie and its target platform are also returned via the result so
// the frontend can persist it into config.json.
func (a *App) CheckQRLogin(source, key string) (model.QRLoginResult, error) {
	fn := core.GetQRLoginCheckFunc(source)
	if fn == nil {
		return model.QRLoginResult{}, fmt.Errorf("unsupported qr login source: %s", source)
	}
	result, err := fn(key)
	if err != nil {
		return model.QRLoginResult{}, err
	}
	if result == nil {
		return model.QRLoginResult{}, fmt.Errorf("empty qr login result")
	}
	if result.Status == model.QRLoginStatusSuccess {
		cookie := qrLoginCookieString(result)
		if cookie != "" {
			cookieSource := qrLoginCookieSource(source)
			result.Cookie = cookie
			core.CM.SetAll(map[string]string{cookieSource: cookie})
			if result.Extra == nil {
				result.Extra = make(map[string]string)
			}
			result.Extra["cookie_saved"] = "true"
			result.Extra["cookie_source"] = cookieSource
			result.Extra["cookie_length"] = strconv.Itoa(len(cookie))
		}
	}
	return *result, nil
}

// qrLoginCookieString flattens a QR login result into a single Cookie header
// string (mirrors the go-music-dl web layer behaviour).
func qrLoginCookieString(result *model.QRLoginResult) string {
	if result == nil {
		return ""
	}
	if cookie := strings.TrimSpace(result.Cookie); cookie != "" {
		return cookie
	}
	if len(result.Cookies) == 0 {
		return ""
	}
	keys := make([]string, 0, len(result.Cookies))
	for k := range result.Cookies {
		if strings.TrimSpace(k) == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		v := strings.TrimSpace(result.Cookies[k])
		if v == "" {
			continue
		}
		parts = append(parts, k+"="+v)
	}
	return strings.Join(parts, "; ")
}

// qrLoginCookieSource maps a QR login source to the platform whose cookie store
// should receive the resolved cookie (WeChat login feeds the QQ platform).
func qrLoginCookieSource(source string) string {
	if source == "qq_wx" {
		return "qq"
	}
	return source
}
