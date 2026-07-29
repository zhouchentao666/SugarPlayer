// Tauri 命令封装层：函数名与原 Wails bindings 保持一致，
// 便于业务代码只改 import 路径即可迁移。
import { invoke } from '@tauri-apps/api/core'
import type { SongMetadata } from '../types'

// ---------- 基础 ----------

export function Version(): Promise<string> {
  return invoke('version')
}

export function AudioServerURL(): Promise<string> {
  return invoke('audio_server_url')
}

export function GetDonateImageURLs(): Promise<Record<string, string>> {
  return invoke('get_donate_image_urls')
}

export function OpenURL(u: string): Promise<void> {
  return invoke('open_url', { u })
}

export function QuitApp(): Promise<void> {
  return invoke('quit_app')
}

// ---------- 本地音乐 ----------

export function ScanMusicFolder(path: string): Promise<string[]> {
  return invoke('scan_music_folder', { path })
}

export function ReadMetadata(path: string): Promise<SongMetadata> {
  return invoke('read_metadata', { path })
}

export function ReadLyrics(path: string): Promise<string> {
  return invoke('read_lyrics', { path })
}

export function ReadCoverArt(path: string): Promise<string> {
  return invoke('read_cover_art', { path })
}

export function ReadImageFile(path: string): Promise<string> {
  return invoke('read_image_file', { path })
}

export function ReadAudioFile(path: string): Promise<string> {
  return invoke('read_audio_file', { path })
}

// ---------- 配置 ----------

export function LoadConfig(): Promise<any> {
  return invoke('load_config')
}

export function SaveConfig(config: any): Promise<void> {
  return invoke('save_config', { config })
}

// ---------- 对话框 / 文件 ----------

export function OpenMusicFiles(): Promise<string[]> {
  return invoke('open_music_files')
}

export function OpenMusicFolder(): Promise<string> {
  return invoke('open_music_folder')
}

export function OpenImageFile(): Promise<string> {
  return invoke('open_image_file')
}

export function OpenInExplorer(path: string): Promise<void> {
  return invoke('open_in_explorer', { path })
}

export function OpenSongEditor(path: string): Promise<void> {
  return invoke('open_song_editor', { path })
}

export function EmitMetadataChanged(): Promise<void> {
  return invoke('emit_metadata_changed')
}

// ---------- 文件夹监听 ----------

export function WatchMusicFolder(path: string): Promise<void> {
  return invoke('watch_music_folder', { path })
}

export function StopWatching(): Promise<void> {
  return invoke('stop_watching')
}

// ---------- 桌面歌词 ----------

export function ToggleDesktopLyric(enabled: boolean): Promise<void> {
  return invoke('toggle_desktop_lyric', { enabled })
}

export function CloseDesktopLyric(): Promise<void> {
  return invoke('close_desktop_lyric')
}

export function GetDesktopLyricConfig(): Promise<string> {
  return invoke('get_desktop_lyric_config')
}

export function SetDesktopLyricBounds(x: number, y: number, width: number, height: number): Promise<void> {
  return invoke('set_desktop_lyric_bounds', { x, y, width, height })
}

export function SetDesktopLyricIgnoreMouseEvents(ignore: boolean): Promise<void> {
  return invoke('set_desktop_lyric_ignore_mouse_events', { ignore })
}

// ---------- 托盘 / 系统 ----------

export function EnableTray(enabled: boolean): Promise<void> {
  return invoke('enable_tray', { enabled })
}

export function SetTraySongInfo(label: string): Promise<void> {
  return invoke('set_tray_song_info', { label })
}

export function SetCloseToTray(enabled: boolean): Promise<void> {
  return invoke('set_close_to_tray', { enabled })
}

export function ShowMainWindow(): Promise<void> {
  return invoke('show_main_window')
}

export function ApplyAutoStart(enabled: boolean): Promise<void> {
  return invoke('apply_auto_start', { enabled })
}
