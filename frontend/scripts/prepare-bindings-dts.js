import { writeFileSync, existsSync, mkdirSync } from 'fs'
import { dirname, join } from 'path'
import { fileURLToPath } from 'url'

const __dirname = dirname(fileURLToPath(import.meta.url))
const bindingsDir = join(__dirname, '..', 'bindings', 'sugarplayer')

if (!existsSync(bindingsDir)) {
  mkdirSync(bindingsDir, { recursive: true })
}

const appDts = `import { CancellablePromise } from "@wailsio/runtime";
import * as models from "./models.js";

export function ApplyAutoStart(enabled: boolean): CancellablePromise<void>;
export function AudioServerURL(): CancellablePromise<string>;
export function CheckUpdate(): CancellablePromise<models.UpdateInfo>;
export function CloseDesktopLyric(): CancellablePromise<void>;
export function EmitMetadataChanged(): CancellablePromise<void>;
export function EnableTray(enabled: boolean): CancellablePromise<void>;
export function GetDesktopLyricConfig(): CancellablePromise<string>;
export function Greet(name: string): CancellablePromise<string>;
export function LoadConfig(): CancellablePromise<models.AppConfig>;
export function OpenImageFile(): CancellablePromise<string>;
export function OpenInExplorer(path: string): CancellablePromise<void>;
export function OpenMusicFiles(): CancellablePromise<string[]>;
export function OpenSongEditor(path: string): CancellablePromise<void>;
export function OpenMusicFolder(): CancellablePromise<string>;
export function OpenURL(u: string): CancellablePromise<void>;
export function ReadAudioFile(path: string): CancellablePromise<string>;
export function ReadCoverArt(path: string): CancellablePromise<string>;
export function ReadImageFile(path: string): CancellablePromise<string>;
export function ReadLyrics(path: string): CancellablePromise<string>;
export function ReadMetadata(path: string): CancellablePromise<models.SongMetadata>;
export function SaveConfig(config: models.AppConfig): CancellablePromise<void>;
export function ScanMusicFolder(path: string): CancellablePromise<string[]>;
export function SetCloseToTray(enabled: boolean): CancellablePromise<void>;
export function SetDesktopLyricBounds(x: number, y: number, width: number, height: number): CancellablePromise<void>;
export function SetDesktopLyricIgnoreMouseEvents(ignore: boolean): CancellablePromise<void>;
export function SetTraySongInfo(label: string): CancellablePromise<void>;
export function ShowMainWindow(): CancellablePromise<void>;
export function StopWatching(): CancellablePromise<void>;
export function ToggleDesktopLyric(enabled: boolean): CancellablePromise<void>;
export function WatchMusicFolder(path: string): CancellablePromise<void>;
export function Version(): CancellablePromise<string>;
export function OnlineSearch(keyword: string, sources: string[]): CancellablePromise<models.OnlineSong[]>;
export function OnlineLyric(song: models.OnlineSong): CancellablePromise<string>;
export function OnlineSources(): CancellablePromise<models.OnlineSource[]>;
export function OnlineVerifyKey(key: string): CancellablePromise<boolean>;
export function OnlineIsUnlocked(): CancellablePromise<boolean>;
export function OnlineDownload(song: models.OnlineSong, opts: models.OnlineDownloadOpts): CancellablePromise<models.OnlineDownloadResult>;
export function GetPlatformCookies(): CancellablePromise<Record<string, string>>;
export function SetPlatformCookies(cookies: Record<string, string>): CancellablePromise<void>;
export function SwitchSongSource(song: models.OnlineSong): CancellablePromise<models.OnlineSong>;
export function QRLoginSources(): CancellablePromise<string[]>;
export function CreateQRLogin(source: string): CancellablePromise<models.QRLoginSession>;
export function CheckQRLogin(source: string, key: string): CancellablePromise<models.QRLoginResult>;
export function OnlineRecommendPlaylists(sources: string[]): CancellablePromise<models.OnlineCollection[]>;
export function OnlineUserPlaylists(sources: string[]): CancellablePromise<models.OnlineCollection[]>;
export function OnlineSearchCollections(keyword: string, kind: string, sources: string[]): CancellablePromise<models.OnlineCollection[]>;
export function OnlineCollectionSongs(collection: models.OnlineCollection): CancellablePromise<models.OnlineSong[]>;
export function OnlineComments(song: models.OnlineSong, kind: string, page: number): CancellablePromise<models.OnlineCommentPage>;
export function OnlinePlaylistCategories(sources: string[]): CancellablePromise<models.OnlineCategorySource[]>;
export function OnlineCategoryPlaylists(source: string, categoryID: string, categoryName: string): CancellablePromise<models.OnlineCollection[]>;
export function OnlineQualityLevels(song: models.OnlineSong): CancellablePromise<string[]>;
`

const modelsDts = `export interface AppConfig {
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
`

writeFileSync(join(bindingsDir, 'app.d.ts'), appDts)
writeFileSync(join(bindingsDir, 'models.d.ts'), modelsDts)
