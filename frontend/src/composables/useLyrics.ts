import { ref, watch, type Ref } from 'vue'
import type { LyricLine } from '@applemusic-like-lyrics/core'
import { parseLrc, parseLrcA2, parseYrc } from '@applemusic-like-lyrics/lyric'
import { ReadLyrics } from '../../bindings/sugarplayer/app'
import {
  convertLrcFormat,
  isA2Format,
  isEnhancedLrc,
  isYrcFormat,
  sanitizeLyricLines,
} from '../utils/lyricConverter'
import { localMetadata } from './useLocalMetadata'

export type { LyricLine }

function normalizeLrcA2(lrc: string): string {
  return lrc
    .replace(/\r\n/g, '\n')
    .replace(/\r/g, '\n')
    .replace(/<00:00\.000>(?=\s*<\d{2}:\d{2}(?:\.\d+)?>)/g, '')
}

function parseLyric(lrc: string): LyricLine[] {
  let parsed: LyricLine[] = []
  if (isYrcFormat(lrc)) {
    parsed = parseYrc(lrc) as LyricLine[]
  } else if (isEnhancedLrc(lrc)) {
    parsed = parseLrcA2(convertLrcFormat(lrc)) as LyricLine[]
  } else if (isA2Format(lrc)) {
    parsed = parseLrcA2(normalizeLrcA2(lrc)) as LyricLine[]
  } else {
    parsed = parseLrc(lrc) as LyricLine[]
  }
  return sanitizeLyricLines(parsed)
}

export function useLyrics(currentSong: Ref<{ path: string } | null>) {
  const lyrics = ref<LyricLine[]>([])
  const hasLyrics = ref(false)

  async function loadLyrics(path: string | null) {
    if (!path) {
      lyrics.value = []
      hasLyrics.value = false
      return
    }

    const override = localMetadata.value[path]?.lyrics
    const lrc = (override && override.trim() !== '') ? override : await ReadLyrics(path).catch(() => '')
    if (!lrc) {
      lyrics.value = []
      hasLyrics.value = false
      return
    }
    const parsed = parseLyric(lrc)
    lyrics.value = parsed
    hasLyrics.value = parsed.length > 0
  }

  watch(currentSong, (song) => {
    loadLyrics(song?.path || null)
  }, { immediate: true })

  watch(() => localMetadata.value[currentSong.value?.path ?? '']?.lyrics, () => {
    loadLyrics(currentSong.value?.path || null)
  })

  return {
    lyrics,
    hasLyrics,
  }
}
