import { ref, computed, nextTick, onMounted, onUnmounted, watch, type Ref } from 'vue'
import { ReadCoverArt, AudioServerURL, OnlineLyric } from '../../bindings/sugarplayer/app'
import { type Song } from '../types'
import type { OnlineSong } from '../../bindings/sugarplayer/models'
import { localMetadata } from './useLocalMetadata'
import { currentOnlineSong } from './onlineState'

interface AudioPlayerOptions {
  audioRef?: Ref<HTMLAudioElement | null>
  onEnded?: () => void
  onOnlinePlayError?: (song: OnlineSong) => void
}

export function useAudioPlayer(options: AudioPlayerOptions = {}) {
  const internalAudioRef = ref<HTMLAudioElement | null>(null)
  const audioRef = options.audioRef || internalAudioRef
  const currentSong = ref<Song | null>(null)
  const isPlaying = ref(false)
  const currentTime = ref(0)
  const duration = ref(0)
  const volume = ref(100)
  const playbackRate = ref(1)
  const coverUrl = ref<string | null>(null)
  const playlistId = ref<string | null>(null)
  const index = ref(-1)
  const serverUrl = ref<string>('')

  // Online playback queue (songs searched from music sources).
  const onlineList = ref<OnlineSong[]>([])
  const onlineIndex = ref(-1)

  const hasSong = computed(() => currentSong.value !== null)

  async function loadCover(path: string) {
    if (path.startsWith('http://') || path.startsWith('https://')) return
    const override = localMetadata.value[path]?.cover
    if (override) {
      coverUrl.value = override
      return
    }
    try {
      coverUrl.value = await ReadCoverArt(path)
    } catch {
      coverUrl.value = null
    }
  }

  watch(() => localMetadata.value[currentSong.value?.path ?? '']?.cover, () => {
    if (currentSong.value) loadCover(currentSong.value.path)
  })

  async function audioUrl(path: string): Promise<string> {
    if (!serverUrl.value) {
      serverUrl.value = await AudioServerURL()
    }
    return `${serverUrl.value}/audio?path=${encodeURIComponent(path)}`
  }

  async function play(playlist: string, songIndex: number, song: Song, autoPlay = true) {
    // Favorited online songs are stored with the online:// scheme so they keep
    // working across app restarts (the streaming port is random per launch).
    if (song.path.startsWith('online://')) {
      playOnlineFavorited(song, playlist, songIndex, autoPlay)
      return
    }
    playlistId.value = playlist
    index.value = songIndex
    currentSong.value = song
    currentTime.value = 0
    duration.value = song.metadata?.duration || 0
    await loadCover(song.path)

    await nextTick()
    if (!audioRef.value) return
    try {
      audioRef.value.src = await audioUrl(song.path)
      audioRef.value.load()
      audioRef.value.playbackRate = playbackRate.value
      if (autoPlay) {
        await audioRef.value.play()
        isPlaying.value = true
      } else {
        isPlaying.value = false
      }
    } catch {
      isPlaying.value = false
    }
  }

  // Play a favorited online song (path uses the online:// scheme) through the
  // bottom bar, reconstructing the streaming URL from the local audio server.
  async function playOnlineFavorited(song: Song, playlist: string, songIndex: number, autoPlay = true) {
    const u = new URL(song.path)
    const source = u.host
    const id = decodeURIComponent(u.pathname.replace(/^\//, ''))
    const extra = u.searchParams.get('extra') || ''
    const onlineSong: OnlineSong = {
      id,
      source,
      name: song.title,
      artist: song.metadata?.artist || '',
      album: song.metadata?.album || '',
      cover: song.cover || '',
      duration: song.metadata?.duration || 0,
      extra,
      link: '',
      streamUrl: '',
    }
    currentOnlineSong.value = onlineSong
    playlistId.value = playlist
    index.value = songIndex
    currentSong.value = song
    currentTime.value = 0
    duration.value = song.metadata?.duration || 0
    coverUrl.value = song.cover || null
    onlineList.value = []
    onlineIndex.value = -1

    await nextTick()
    if (!audioRef.value) return
    try {
      const base = serverUrl.value || (await AudioServerURL())
      serverUrl.value = base
      const q = new URLSearchParams({ source, id })
      if (extra) q.set('extra', extra)
      audioRef.value.src = `${base}/online?${q.toString()}`
      audioRef.value.load()
      audioRef.value.playbackRate = playbackRate.value
      if (autoPlay) {
        await audioRef.value.play()
        isPlaying.value = true
      } else {
        isPlaying.value = false
      }
    } catch {
      isPlaying.value = false
    }
  }

  function togglePlay() {
    if (!currentSong.value || !audioRef.value) return
    if (isPlaying.value) {
      audioRef.value.pause()
    } else {
      audioRef.value.play().catch(() => {})
    }
  }

  // Play an online song (searched from a music source) through the bottom bar.
  async function playOnline(song: OnlineSong, list: OnlineSong[], songIndex: number, autoPlay = true) {
    currentOnlineSong.value = song
    const adapted: Song = {
      id: `online:${song.source}:${song.id}`,
      path: song.streamUrl,
      title: song.name,
      metadata: {
        title: song.name,
        artist: song.artist,
        album: song.album,
        genre: '',
        year: '',
        duration: song.duration,
        bitrate: 0,
      },
    }
    playlistId.value = 'online'
    index.value = songIndex
    currentSong.value = adapted
    currentTime.value = 0
    duration.value = song.duration || 0
    coverUrl.value = song.cover || null
    onlineList.value = list
    onlineIndex.value = songIndex

    await nextTick()
    if (!audioRef.value) return
    try {
      audioRef.value.src = song.streamUrl
      audioRef.value.load()
      audioRef.value.playbackRate = playbackRate.value
      if (autoPlay) {
        await audioRef.value.play()
        isPlaying.value = true
      } else {
        isPlaying.value = false
      }
    } catch {
      isPlaying.value = false
    }
  }

  function pause() {
    audioRef.value?.pause()
  }

  function seek(time: number) {
    if (!audioRef.value) return
    audioRef.value.currentTime = time
    currentTime.value = time
  }

  function setVolume(value: number) {
    volume.value = value
    if (audioRef.value) audioRef.value.volume = value / 100
  }

  function setPlaybackRate(rate: number) {
    const clamped = Math.min(16, Math.max(0.25, rate))
    playbackRate.value = clamped
    if (audioRef.value) audioRef.value.playbackRate = clamped
  }

  function bindAudioEvents() {
    const audio = audioRef.value
    if (!audio) return
    audio.volume = volume.value / 100
    audio.addEventListener('timeupdate', () => {
      currentTime.value = audio.currentTime || 0
    })
    audio.addEventListener('loadedmetadata', () => {
      duration.value = audio.duration || currentSong.value?.metadata?.duration || 0
    })
    if (options.onEnded) {
      audio.addEventListener('ended', options.onEnded)
    }
    audio.addEventListener('play', () => { isPlaying.value = true })
    audio.addEventListener('pause', () => { isPlaying.value = false })
    audio.addEventListener('error', () => {
      if (options.onOnlinePlayError && currentOnlineSong.value) {
        options.onOnlinePlayError(currentOnlineSong.value)
      }
    })
  }

  onMounted(() => {
    nextTick(bindAudioEvents)
  })

  onUnmounted(() => {
    audioRef.value?.pause()
  })

  return {
    audioRef,
    currentSong,
    isPlaying,
    currentTime,
    duration,
    volume,
    playbackRate,
    coverUrl,
    playlistId,
    index,
    hasSong,
    onlineList,
    onlineIndex,
    play,
    playOnline,
    togglePlay,
    pause,
    seek,
    setVolume,
    setPlaybackRate,
  }
}
