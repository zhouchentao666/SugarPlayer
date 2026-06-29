import { CancellablePromise } from "@wailsio/runtime";
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
