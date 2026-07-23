package migu

import "sugarplayer/internal/music/model"

func GetUserPlaylists(page, limit int) ([]model.Playlist, error) {
	return defaultMigu.GetUserPlaylists(page, limit)
}

func (p *Migu) GetUserPlaylists(page, limit int) ([]model.Playlist, error) {
	return nil, model.ErrUserPlaylistsUnsupported
}
