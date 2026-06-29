import { ref, computed, nextTick, onMounted, onUnmounted, type Ref } from 'vue'
import { ReadCoverArt, AudioServerURL } from '../../bindings/sugarplayer/app'
import { type Song } from '../types'

interface AudioPlayerOptions {
  audioRef?: Ref<HTMLAudioElement | null>
  onEnded?: () => void
}

export function useAudioPlayer(options: AudioPlayerOptions = {}) {
  const internalAudioRef = ref<HTMLAudioElement | null>(null)
  const audioRef = options.audioRef || internalAudioRef
  const currentSong = ref<Song | null>(null)
  const isPlaying = ref(false)
  const currentTime = ref(0)
  const duration = ref(0)
  const volume = ref(100)
  const coverUrl = ref<string | null>(null)
  const playlistId = ref<string | null>(null)
  const index = ref(-1)
  const serverUrl = ref<string>('')

  const hasSong = computed(() => currentSong.value !== null)

  async function loadCover(path: string) {
    try {
      coverUrl.value = await ReadCoverArt(path)
    } catch {
      coverUrl.value = null
    }
  }

  async function audioUrl(path: string): Promise<string> {
    if (!serverUrl.value) {
      serverUrl.value = await AudioServerURL()
    }
    return `${serverUrl.value}/audio?path=${encodeURIComponent(path)}`
  }

  async function play(playlist: string, songIndex: number, song: Song) {
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
      await audioRef.value.play()
      isPlaying.value = true
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
    coverUrl,
    playlistId,
    index,
    hasSong,
    play,
    togglePlay,
    pause,
    seek,
    setVolume,
  }
}
