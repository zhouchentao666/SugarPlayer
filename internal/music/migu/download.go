package migu

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sugarplayer/internal/music/model"
)

func GetDownloadURL(s *model.Song) (string, error) { return defaultMigu.GetDownloadURL(s) }

// GetDownloadURL 获取下载链接
func (m *Migu) GetDownloadURL(s *model.Song) (string, error) {
	if s.Source != "migu" {
		return "", errors.New("source mismatch")
	}
	if s.URL != "" {
		return s.URL, nil
	}

	var contentID, resourceType, formatType string
	if s.Extra != nil {
		contentID = s.Extra["content_id"]
		resourceType = s.Extra["resource_type"]
		formatType = s.Extra["format_type"]
	}

	if contentID == "" || resourceType == "" || formatType == "" {
		parts := strings.Split(s.ID, "|")
		if len(parts) == 3 {
			contentID = parts[0]
			resourceType = parts[1]
			formatType = parts[2]
		} else {
			return "", errors.New("invalid id structure and missing extra data")
		}
	}

	// 尝试获取播放链接（优先使用新 API）
	if url := m.getPlayUrl(contentID, resourceType, formatType); url != "" {
		return url, nil
	}

	// 备用：使用旧版 listenSong API
	return m.getListenSongUrl(contentID, resourceType, formatType)
}

// getPlayUrl 使用新版 API 获取播放链接
func (m *Migu) getPlayUrl(contentID, resourceType, formatType string) string {
	params := url.Values{}
	params.Set("copyrightId", "0")
	params.Set("contentId", contentID)
	params.Set("resourceType", resourceType)
	params.Set("toneFlag", formatType)
	params.Set("userId", MagicUserID)
	params.Set("channel", "0")

	apiURL := "https://app.pd.nf.migu.cn/MIGUM3.0/v1.0/content/sub/listenSong.do?" + params.Encode()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 10; SM-G981B) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://app.pd.nf.migu.cn/")
	req.Header.Set("Cookie", m.cookie)

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode == 302 {
		location := resp.Header.Get("Location")
		if location != "" && !strings.Contains(location, "error") {
			return location
		}
	}
	return ""
}

// getListenSongUrl 使用旧版 API 获取播放链接
func (m *Migu) getListenSongUrl(contentID, resourceType, formatType string) (string, error) {
	params := url.Values{}
	params.Set("toneFlag", formatType)
	params.Set("netType", "00")
	params.Set("userId", MagicUserID)
	params.Set("ua", "Android_migu")
	params.Set("version", "5.1")
	params.Set("copyrightId", "0")
	params.Set("contentId", contentID)
	params.Set("resourceType", resourceType)
	params.Set("channel", "0")

	apiURL := "http://app.pd.nf.migu.cn/MIGUM2.0/v1.0/content/sub/listenSong.do?" + params.Encode()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Referer", Referer)
	req.Header.Set("Cookie", m.cookie)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 302 {
		location := resp.Header.Get("Location")
		if location != "" && !strings.Contains(location, "error") {
			return location, nil
		}
	}

	return "", errors.New("migu download url not found")
}
