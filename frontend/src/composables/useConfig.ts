import { ref, watch, type Ref } from 'vue'
import { SaveConfig, LoadConfig } from '../../bindings/sugarplayer/app'
import { type Playlist } from '../types'
import type { SortMode, SortOrder } from './usePlaylistView'
import type { LocalSongMetadata } from './useLocalMetadata'

export type WindowEffect = 'none' | 'acrylic' | 'custom-image' | 'song-color'
export type FullScreenBackground = 'static' | 'dynamic'
export type CoverTransition = 'fade' | 'slide-left' | 'slide-both'
export type HotkeyAction = 'togglePlay' | 'prevSong' | 'nextSong' | 'volumeUp' | 'volumeDown' | 'mute' | 'togglePlayerDetail'

export const HOTKEY_ACTIONS: { value: HotkeyAction; label: string }[] = [
  { value: 'togglePlay', label: '播放/暂停' },
  { value: 'prevSong', label: '上一首' },
  { value: 'nextSong', label: '下一首' },
  { value: 'volumeUp', label: '增大音量' },
  { value: 'volumeDown', label: '减小音量' },
  { value: 'mute', label: '静音' },
  { value: 'togglePlayerDetail', label: '全屏播放器' },
]

export const DEFAULT_HOTKEYS: Record<HotkeyAction, string> = {
  togglePlay: ' ',
  prevSong: 'ArrowLeft',
  nextSong: 'ArrowRight',
  volumeUp: 'ArrowUp',
  volumeDown: 'ArrowDown',
  mute: 'm',
  togglePlayerDetail: 'i',
}

export interface PlaylistSort {
  mode: SortMode
  order: SortOrder
}

export interface AppSettings {
  theme: 'system' | 'light' | 'dark'
  accentColor: string
  quality: 'standard' | 'high' | 'lossless'
  autoplay: boolean
  savePlaylistAndSong: boolean
  saveWindowPosition: boolean
  windowEffect: WindowEffect
  customImagePath: string
  customImageOpacity: number
  customImageBlur: number
  songColorOpacity: number
  songColorBlur: number
  fullScreenBackground: FullScreenBackground
  coverTransition: CoverTransition
  immersivePlayerBar: boolean
  hotkeys: Partial<Record<HotkeyAction, string>>
  checkUpdateOnStartup: boolean
  selectedPlaylistId: string
  playlistSorts: Record<string, PlaylistSort>
  localMetadata: Record<string, LocalSongMetadata>
}

export interface ConfigPlayback {
  playlistId: string
  songIndex: number
  time: number
}

export interface ConfigWindow {
  x: number
  y: number
  width: number
  height: number
}

export function useConfig(
  playlists: Ref<Playlist[]>,
  settings: Ref<AppSettings>,
  playback: Ref<ConfigPlayback>,
  windowState: Ref<ConfigWindow>,
  isLoading: Ref<boolean>
) {
  function buildConfig() {
    return {
      playlists: playlists.value,
      settings: settings.value,
      playback: playback.value,
      window: windowState.value,
    }
  }

  async function save() {
    if (isLoading.value) return
    await SaveConfig(buildConfig())
  }

  async function load() {
    try {
      const config = await LoadConfig()
      if (config.playlists && config.playlists.length > 0) {
        playlists.value = config.playlists as Playlist[]
      }
      if (config.settings) {
        const hasEffect = Boolean(config.settings.windowEffect)
        settings.value = {
          theme: (config.settings.theme as AppSettings['theme']) || 'system',
          accentColor: config.settings.accentColor || '#0078d4',
          quality: (config.settings.quality as AppSettings['quality']) || 'standard',
          autoplay: config.settings.autoplay ?? false,
          savePlaylistAndSong: config.settings.savePlaylistAndSong ?? true,
          saveWindowPosition: config.settings.saveWindowPosition ?? true,
          windowEffect: (config.settings.windowEffect as WindowEffect) || 'acrylic',
          customImagePath: config.settings.customImagePath || '',
          customImageOpacity: hasEffect ? (config.settings.customImageOpacity ?? 35) : 35,
          customImageBlur: hasEffect ? (config.settings.customImageBlur ?? 20) : 20,
          songColorOpacity: hasEffect ? (config.settings.songColorOpacity ?? 45) : 45,
          songColorBlur: hasEffect ? (config.settings.songColorBlur ?? 30) : 30,
          fullScreenBackground: (config.settings.fullScreenBackground as FullScreenBackground) || 'static',
          coverTransition: ((config.settings as unknown as Record<string, unknown>).coverTransition as CoverTransition) || 'fade',
          immersivePlayerBar: config.settings.immersivePlayerBar ?? false,
          hotkeys: ((config.settings as unknown as Record<string, unknown>).hotkeys as Record<string, string>) || { ...DEFAULT_HOTKEYS },
          checkUpdateOnStartup: ((config.settings as unknown as Record<string, unknown>).checkUpdateOnStartup as boolean) ?? true,
          selectedPlaylistId: config.settings.selectedPlaylistId ?? '',
          playlistSorts: (config.settings.playlistSorts as Record<string, PlaylistSort>) ?? {},
          localMetadata: (config.settings.localMetadata as Record<string, LocalSongMetadata>) ?? {},
        }
      }
      if (config.playback) {
        playback.value = {
          playlistId: config.playback.playlistId || '',
          songIndex: config.playback.songIndex ?? -1,
          time: config.playback.time ?? 0,
        }
      }
      if (config.window) {
        windowState.value = {
          x: config.window.x ?? 0,
          y: config.window.y ?? 0,
          width: config.window.width ?? 800,
          height: config.window.height ?? 600,
        }
      }
    } catch {
      // 首次启动没有配置文件
    } finally {
      isLoading.value = false
    }
  }

  watch(playlists, save, { deep: true })
  watch(settings, save, { deep: true })

  return { save, load }
}
