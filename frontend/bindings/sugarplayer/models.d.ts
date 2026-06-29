export interface AppConfig {
  playlists: ConfigPlaylist[] | null;
  settings: ConfigSettings;
}

export interface ConfigPlaylist {
  id: string;
  name: string;
  songs: ConfigSong[] | null;
  folders: string[] | null;
}

export interface ConfigSettings {
  theme: string;
  accentColor: string;
  quality: string;
  autoplay: boolean;
  windowEffect: string;
  customImagePath: string;
  customImageOpacity: number;
  customImageBlur: number;
  songColorOpacity: number;
  songColorBlur: number;
}

export interface ConfigSong {
  id: string;
  path: string;
  title: string;
  metadata?: SongMetadata | null;
}

export interface SongMetadata {
  title: string;
  artist: string;
  album: string;
  genre: string;
  year: string;
  duration: number;
  bitrate: number;
}
