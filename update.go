package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	lanzouURL        = "https://wwazq.lanzoub.com/b00oe00xsb"
	lanzouPassword   = "1234"
	githubRepo       = "zhouchentao666/SugarPlayer"
	githubReleaseURL = "https://github.com/zhouchentao666/SugarPlayer/releases"
)

// UpdateInfo holds the result of an update check.
type UpdateInfo struct {
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	HasUpdate      bool   `json:"hasUpdate"`
	ReleaseURL     string `json:"releaseUrl"`
	LanzouURL      string `json:"lanzouUrl"`
	LanzouPassword string `json:"lanzouPassword"`
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

// CheckUpdate checks GitHub for a newer release. It tries the API first, then
// the release page, then a mirror.
func (a *App) CheckUpdate() (UpdateInfo, error) {
	latest, releaseURL, err := fetchLatestVersion()
	info := UpdateInfo{
		CurrentVersion: a.Version(),
		LatestVersion:  latest,
		ReleaseURL:     releaseURL,
		LanzouURL:      lanzouURL,
		LanzouPassword: lanzouPassword,
	}
	if err != nil {
		info.HasUpdate = false
		return info, err
	}
	cmp, err := compareVersions(a.Version(), latest)
	if err != nil {
		info.HasUpdate = false
		return info, err
	}
	info.HasUpdate = cmp < 0
	return info, nil
}

func fetchLatestVersion() (version string, releaseURL string, err error) {
	version, releaseURL, err = fetchFromGitHubAPI()
	if err == nil && version != "" {
		return version, releaseURL, nil
	}

	version, releaseURL, err = fetchFromGitHubPage()
	if err == nil && version != "" {
		return version, releaseURL, nil
	}

	return fetchFromMirror()
}

func httpClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Second}
}

func fetchFromGitHubAPI() (string, string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "SugarPlayer-UpdateChecker")

	resp, err := httpClient().Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("github api returned %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", err
	}
	if release.TagName == "" {
		return "", "", errors.New("empty tag name")
	}
	if release.HTMLURL == "" {
		release.HTMLURL = githubReleaseURL
	}
	return normalizeVersion(release.TagName), release.HTMLURL, nil
}

var releaseTagRe = regexp.MustCompile(`/releases/tag/([^"\s]+)`)

func fetchFromGitHubPage() (string, string, error) {
	req, err := http.NewRequest("GET", githubReleaseURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "SugarPlayer-UpdateChecker")

	resp, err := httpClient().Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("github page returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	matches := releaseTagRe.FindAllStringSubmatch(string(body), -1)
	if len(matches) == 0 {
		return "", "", errors.New("no release tags found on page")
	}

	best := matches[0][1]
	for _, m := range matches {
		v := m[1]
		if cmp, _ := compareVersions(best, v); cmp < 0 {
			best = v
		}
	}
	releaseURL := fmt.Sprintf("%s/tag/%s", githubReleaseURL, best)
	return normalizeVersion(best), releaseURL, nil
}

func fetchFromMirror() (string, string, error) {
	mirrors := []string{
		"https://ghproxy.com/https://api.github.com/repos/" + githubRepo + "/releases/latest",
		"https://mirror.ghproxy.com/https://api.github.com/repos/" + githubRepo + "/releases/latest",
		"https://gh.api.99988866.xyz/https://api.github.com/repos/" + githubRepo + "/releases/latest",
	}

	for _, url := range mirrors {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("User-Agent", "SugarPlayer-UpdateChecker")

		resp, err := httpClient().Do(req)
		if err != nil {
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			continue
		}

		var release githubRelease
		if err := json.Unmarshal(body, &release); err != nil {
			continue
		}
		if release.TagName == "" {
			continue
		}
		if release.HTMLURL == "" {
			release.HTMLURL = githubReleaseURL
		}
		return normalizeVersion(release.TagName), release.HTMLURL, nil
	}

	return "", "", errors.New("all update sources failed")
}

func normalizeVersion(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "v")
}

func parseVersion(v string) ([]int, error) {
	v = normalizeVersion(v)
	parts := strings.Split(v, ".")
	if len(parts) < 2 || len(parts) > 4 {
		return nil, fmt.Errorf("unsupported version format: %s", v)
	}
	nums := make([]int, 4)
	for i := 0; i < 4; i++ {
		if i < len(parts) {
			n, err := strconv.Atoi(parts[i])
			if err != nil {
				return nil, fmt.Errorf("invalid version segment: %s", parts[i])
			}
			nums[i] = n
		}
	}
	return nums, nil
}

func compareVersions(a, b string) (int, error) {
	av, err := parseVersion(a)
	if err != nil {
		return 0, err
	}
	bv, err := parseVersion(b)
	if err != nil {
		return 0, err
	}
	for i := 0; i < 4; i++ {
		if av[i] < bv[i] {
			return -1, nil
		}
		if av[i] > bv[i] {
			return 1, nil
		}
	}
	return 0, nil
}
