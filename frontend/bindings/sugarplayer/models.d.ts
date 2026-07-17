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

export interface ConfigDesktopLyric {
  enabled: boolean;
  fontSize: number;
  mainColor: string;
  unplayedColor: string;
  shadowColor: string;
  fontWeight: number;
  position: string;
  alwaysShowPlayInfo: boolean;
  animation: boolean;
  showYrc: boolean;
  showTran: boolean;
  isDoubleLine: boolean;
  textBackgroundMask: boolean;
  backgroundMaskColor: string;
  fontFamily: string;
  x: number;
  y: number;
  width: number;
  height: number;
  isLock: boolean;
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
  autoStart: boolean;
  trayEnabled: boolean;
  closeToTray: boolean;
  desktopLyric: ConfigDesktopLyric;
  selectedPlaylistId: string;
  playlistSorts: Record<string, ConfigPlaylistSort>;
  localMetadata: Record<string, ConfigLocalMetadata>;
  platformCookies: Record<string, string> | null;
  autoSwitchInvalidSource: boolean;
  pinnedOnlinePlaylists: OnlineCollection[] | null;
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
  cover?: string | null;
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

export interface OnlineSong {
  id: string;
  name: string;
  artist: string;
  album: string;
  cover: string;
  duration: number;
  source: string;
  extra: string;
  link: string;
  streamUrl: string;
}

export interface OnlineSource {
  id: string;
  name: string;
  enabled: boolean;
  recommend: boolean;
  userPlaylists: boolean;
}

export interface OnlineDownloadOpts {
  dir: string;
  withLyrics: boolean;
  withCover: boolean;
  embed: boolean;
  quality: string;
}

export interface OnlineDownloadResult {
  path: string;
  lyricPath: string;
  coverPath: string;
  warning: string;
}

export interface QRLoginSession {
  source: string;
  key: string;
  url: string;
  image_url: string;
  state?: string;
  expires_at?: number;
  extra?: Record<string, string> | null;
}

export interface QRLoginResult {
  source: string;
  key: string;
  status: string;
  message?: string;
  cookie?: string;
  cookies?: Record<string, string> | null;
  extra?: Record<string, string> | null;
}

export interface OnlineCollection {
  id: string;
  name: string;
  cover: string;
  source: string;
  link: string;
  kind: string;
  creator: string;
  trackCount: number;
  extra: string;
}

export interface OnlineCategoryItem {
  id: string;
  name: string;
  hot: boolean;
  source: string;
}

export interface OnlineCategoryGroup {
  name: string;
  categories: OnlineCategoryItem[];
}

export interface OnlineCategorySource {
  source: string;
  name: string;
  groups: OnlineCategoryGroup[];
}

export interface OnlineComment {
  id: string;
  text: string;
  time: number;
  userName: string;
  avatar: string;
  userId: string;
  likedCount: number;
  location: string;
  images: string[];
  replyNum: number;
  reply: OnlineComment[];
}

export interface OnlineCommentPage {
  source: string;
  kind: string;
  comments: OnlineComment[];
  total: number;
  page: number;
  limit: number;
  maxPage: number;
}
