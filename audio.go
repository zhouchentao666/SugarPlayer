package main

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"go.senan.xyz/taglib"
)

// AudioServer serves local audio files over HTTP so the WebView can stream them.
type AudioServer struct {
	server *http.Server
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
	go func() {
		_ = s.server.Serve(listener)
	}()
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
