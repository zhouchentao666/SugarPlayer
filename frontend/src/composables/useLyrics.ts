import { ref, watch, type Ref } from 'vue'
import type { LyricLine } from '@applemusic-like-lyrics/core'
import { parseLrc } from '@applemusic-like-lyrics/lyric'
import { ReadLyrics } from '../../bindings/sugarplayer/app'

export type { LyricLine }

export function useLyrics(currentSong: Ref<{ path: string } | null>) {
  const lyrics = ref<LyricLine[]>([])
  const hasLyrics = ref(false)

  async function loadLyrics(path: string | null) {
    if (!path) {
      lyrics.value = []
      hasLyrics.value = false
      return
    }

    try {
      const lrc = await ReadLyrics(path)
      const parsed = parseLrc(lrc) as LyricLine[]
      lyrics.value = parsed
      hasLyrics.value = parsed.length > 0
    } catch {
      lyrics.value = []
      hasLyrics.value = false
    }
  }

  watch(currentSong, (song) => {
    loadLyrics(song?.path || null)
  }, { immediate: true })

  return {
    lyrics,
    hasLyrics,
  }
}
