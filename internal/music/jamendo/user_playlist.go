package jamendo

import "sugarplayer/internal/music/model"

func GetUserPlaylists(page, limit int) ([]model.Playlist, error) {
	return defaultJamendo.GetUserPlaylists(page, limit)
}

func (p *Jamendo) GetUserPlaylists(page, limit int) ([]model.Playlist, error) {
	return nil, model.ErrUserPlaylistsUnsupported
}
