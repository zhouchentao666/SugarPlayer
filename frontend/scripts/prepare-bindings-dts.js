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

export function AudioServerURL(): CancellablePromise<string>;
export function Greet(name: string): CancellablePromise<string>;
export function LoadConfig(): CancellablePromise<models.AppConfig>;
export function OpenImageFile(): CancellablePromise<string>;
export function OpenMusicFiles(): CancellablePromise<string[]>;
export function OpenMusicFolder(): CancellablePromise<string>;
export function ReadAudioFile(path: string): CancellablePromise<string>;
export function ReadCoverArt(path: string): CancellablePromise<string>;
export function ReadImageFile(path: string): CancellablePromise<string>;
export function ReadLyrics(path: string): CancellablePromise<string>;
export function ReadMetadata(path: string): CancellablePromise<models.SongMetadata>;
export function SaveConfig(config: models.AppConfig): CancellablePromise<void>;
export function ScanMusicFolder(path: string): CancellablePromise<string[]>;
export function StopWatching(): CancellablePromise<void>;
export function WatchMusicFolder(path: string): CancellablePromise<void>;
`

const modelsDts = `export interface AppConfig {
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
`

writeFileSync(join(bindingsDir, 'app.d.ts'), appDts)
writeFileSync(join(bindingsDir, 'models.d.ts'), modelsDts)
