export interface AppConfig {
  playlists: ConfigPlaylist[] | null;
  settings: ConfigSettings;
  playback: ConfigPlayback;
  window: ConfigWindow;
}

export interface ConfigPlaylist {
  id: string;
  name: string;
  songs: ConfigSong[] | null;
  folders: string[] | null;
}

export interface ConfigPlaylistSort {
  mode: string;
  order: string;
}

export interface ConfigLocalMetadata {
  title?: string;
  artist?: string;
  album?: string;
  cover?: string;
  lyrics?: string;
  lyricsFormat?: string;
}

export interface ConfigSettings {
  theme: string;
  accentColor: string;
  quality: string;
  autoplay: boolean;
  savePlaylistAndSong: boolean;
  saveWindowPosition: boolean;
  windowEffect: string;
  customImagePath: string;
  customImageOpacity: number;
  customImageBlur: number;
  songColorOpacity: number;
  songColorBlur: number;
  fullScreenBackground: string;
  immersivePlayerBar: boolean;
  hotkeys: Record<string, string>;
  checkUpdateOnStartup: boolean;
  selectedPlaylistId: string;
  playlistSorts: Record<string, ConfigPlaylistSort>;
  localMetadata: Record<string, ConfigLocalMetadata>;
}

export interface UpdateInfo {
  currentVersion: string;
  latestVersion: string;
  hasUpdate: boolean;
  releaseUrl: string;
  lanzouUrl: string;
  lanzouPassword: string;
}

export interface ConfigPlayback {
  playlistId: string;
  songIndex: number;
  time: number;
}

export interface ConfigWindow {
  x: number;
  y: number;
  width: number;
  height: number;
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
