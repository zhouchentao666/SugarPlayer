package main

import (
	"embed"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"go.senan.xyz/taglib"
)

// donateAssets embeds the donation QR codes so they ship inside the binary.
//
//go:embed assets/微信.jpg assets/支付宝.jpg
var donateAssets embed.FS

// AudioServer serves local audio files over HTTP so the WebView can stream them.
type AudioServer struct {
	server *http.Server
	mux    *http.ServeMux
	port   int
}

func newAudioServer() *AudioServer {
	s := &AudioServer{}
	s.start()
	return s
}

func (s *AudioServer) start() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return
	}
	s.port = listener.Addr().(*net.TCPAddr).Port

	mux := http.NewServeMux()
	s.mux = mux
	mux.HandleFunc("/audio", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Query().Get("path")
		if path == "" || !isAudioFile(path) {
			http.NotFound(w, r)
			return
		}
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.ServeFile(w, r, path)
	})

	s.server = &http.Server{Handler: mux}
	s.registerOnlineProxy()
	s.registerCoverProxy()
	s.registerDonateProxy()
	go func() {
		_ = s.server.Serve(listener)
	}()
}

// registerDonateProxy serves the embedded donation QR codes over the local
// audio server so the sidebar can display them inside the WebView.
func (s *AudioServer) registerDonateProxy() {
	s.mux.HandleFunc("/donate", func(w http.ResponseWriter, r *http.Request) {
		var file string
		switch strings.TrimSpace(r.URL.Query().Get("name")) {
		case "wechat":
			file = "assets/微信.jpg"
		case "alipay":
			file = "assets/支付宝.jpg"
		default:
			http.NotFound(w, r)
			return
		}
		data, err := donateAssets.ReadFile(file)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})
}

// GetDonateImageURLs returns the local server URLs for the donation QR codes.
func (a *App) GetDonateImageURLs() map[string]string {
	base := fmt.Sprintf("http://127.0.0.1:%d", a.audio.port)
	return map[string]string{
		"wechat": base + "/donate?name=wechat",
		"alipay": base + "/donate?name=alipay",
	}
}

func isAudioFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".mp3", ".flac", ".wav", ".aac", ".ogg", ".m4a", ".wma", ".opus":
		return true
	}
	return false
}

// AudioServerURL returns the local audio streaming server URL.
func (a *App) AudioServerURL() string {
	return fmt.Sprintf("http://127.0.0.1:%d", a.audio.port)
}

func first(values []string) string {
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

// ReadMetadata reads metadata from an audio file.
func (a *App) ReadMetadata(path string) (SongMetadata, error) {
	tags, err := taglib.ReadTags(path)
	if err != nil {
		return SongMetadata{}, err
	}
	props, err := taglib.ReadProperties(path)
	if err != nil {
		return SongMetadata{}, err
	}

	return SongMetadata{
		Title:    first(tags[taglib.Title]),
		Artist:   first(tags[taglib.Artist]),
		Album:    first(tags[taglib.Album]),
		Genre:    first(tags[taglib.Genre]),
		Year:     first(tags[taglib.Date]),
		Duration: props.Length.Seconds(),
		Bitrate:  props.Bitrate,
	}, nil
}

// ReadAudioFile reads the raw bytes of an audio file.
func (a *App) ReadAudioFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// ReadCoverArt reads embedded cover art from an audio file and returns a data URL.
func (a *App) ReadCoverArt(path string) (string, error) {
	img, err := taglib.ReadImage(path)
	if err != nil {
		return "", err
	}
	if len(img) == 0 {
		return "", fmt.Errorf("no cover art found")
	}
	mime := http.DetectContentType(img)
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(img), nil
}

// ReadLyrics reads lyrics from an audio file's LYRICS tag or a matching .lrc file.
func (a *App) ReadLyrics(path string) (string, error) {
	tags, err := taglib.ReadTags(path)
	if err == nil {
		if lyrics := first(tags["LYRICS"]); lyrics != "" {
			return lyrics, nil
		}
	}

	dir := filepath.Dir(path)
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	lrcPath := filepath.Join(dir, base+".lrc")
	data, err := os.ReadFile(lrcPath)
	if err != nil {
		return "", fmt.Errorf("no lyrics found")
	}
	return string(data), nil
}
