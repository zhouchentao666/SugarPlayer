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

// 在线歌曲在队列中以 Song 形式存储，id 与 onlineMap 的 key 保持一致
function onlineKey(song: OnlineSong): string {
  return `online:${song.source}:${song.id}`
}

function toAdaptedSong(song: OnlineSong): Song {
  return {
    id: onlineKey(song),
    path: song.streamUrl,
    title: song.name,
    cover: song.cover,
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
  // 播放列表（与歌单解耦的播放队列）
  const queue = ref<Song[]>([])
  const index = ref(-1)
  // 上下文标签（用于展示/恢复），不再用于队列导航
  const playlistId = ref<string | null>(null)
  const serverUrl = ref<string>('')

  // 在线播放队列（搜索/歌单来的在线歌曲），用于换源与在线上下文
  const onlineList = ref<OnlineSong[]>([])
  const onlineIndex = ref(-1)
  // 在线歌曲原始对象映射（queue 中 Song.id -> OnlineSong），换源/重播时使用
  const onlineMap = new Map<string, OnlineSong>()

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

  // 播放本地歌曲（队列中的普通 Song）
  async function playLocal(song: Song, autoPlay = true) {
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

  // 播放已收藏的在线歌曲（path 使用 online:// 方案），通过本地音频服务重建流地址
  async function playOnlineFavorited(song: Song, context: string, songIndex: number, autoPlay = true) {
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
    playlistId.value = context
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

  // 直接用流地址播放在线歌曲（queue 中的在线条目走这里）
  async function playOnlineStream(song: OnlineSong, autoPlay = true) {
    const adapted = toAdaptedSong(song)
    currentSong.value = adapted
    currentTime.value = 0
    duration.value = song.duration || 0
    coverUrl.value = song.cover || null
    playlistId.value = 'online'
    onlineIndex.value = index.value
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

  // 播放队列中第 i 首（根据条目类型分流）
  async function playQueueAt(i: number, autoPlay = true) {
    const song = queue.value[i]
    if (!song) return
    index.value = i
    const om = onlineMap.get(song.id)
    if (om) {
      currentOnlineSong.value = om
      await playOnlineStream(om, autoPlay)
    } else if (song.path.startsWith('online://')) {
      await playOnlineFavorited(song, playlistId.value ?? '', i, autoPlay)
    } else {
      currentOnlineSong.value = null
      await playLocal(song, autoPlay)
    }
  }

  // 替换播放列表：用 songs 替换整个队列并从 startIndex 开始播放
  async function playSongs(songs: Song[], startIndex: number, context?: string | null, autoPlay = true) {
    queue.value = songs
    playlistId.value = context ?? null
    onlineMap.clear()
    onlineList.value = []
    onlineIndex.value = -1
    currentOnlineSong.value = null
    await playQueueAt(startIndex, autoPlay)
  }

  // 播放在线歌曲列表（搜索/歌单结果），整体写入队列并支持换源
  async function playOnline(list: OnlineSong[], startIndex: number, autoPlay = true) {
    onlineList.value = list
    onlineIndex.value = startIndex
    onlineMap.clear()
    for (const o of list) onlineMap.set(onlineKey(o), o)
    queue.value = list.map(toAdaptedSong)
    playlistId.value = 'online'
    await playQueueAt(startIndex, autoPlay)
  }

  // 添加到播放列表（本地/已收藏）：追加到队尾，若当前未播放则立即播放
  function addToQueue(song: Song) {
    if (queue.value.length && queue.value[queue.value.length - 1]?.id === song.id) return
    queue.value = [...queue.value, song]
    if (index.value < 0 && !currentSong.value) {
      playQueueAt(queue.value.length - 1, true)
    }
  }

  // 添加在线歌曲到播放列表
  async function addOnlineToQueue(song: OnlineSong) {
    const adapted = toAdaptedSong(song)
    if (queue.value.length && queue.value[queue.value.length - 1]?.id === adapted.id) return
    onlineMap.set(onlineKey(song), song)
    queue.value = [...queue.value, adapted]
    if (index.value < 0 && !currentSong.value) {
      await playQueueAt(queue.value.length - 1, true)
    }
  }

  // 从播放列表删除第 i 首，自动维护 index
  function removeFromQueue(i: number) {
    if (i < 0 || i >= queue.value.length) return
    const wasCurrent = i === index.value
    queue.value = queue.value.filter((_, idx) => idx !== i)
    if (i < index.value) {
      index.value = index.value - 1
    } else if (wasCurrent) {
      if (queue.value.length === 0) {
        index.value = -1
        currentSong.value = null
        isPlaying.value = false
        audioRef.value?.pause()
      } else {
        const nextIdx = Math.min(i, queue.value.length - 1)
        playQueueAt(nextIdx, isPlaying.value)
      }
    }
  }

  function clearQueue() {
    queue.value = []
    index.value = -1
    currentSong.value = null
    isPlaying.value = false
    onlineMap.clear()
    onlineList.value = []
    onlineIndex.value = -1
    currentOnlineSong.value = null
    audioRef.value?.pause()
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

  function setPlaybackRate(rate: number) {
    const clamped = Math.min(16, Math.max(0.25, rate))
    playbackRate.value = clamped
    if (audioRef.value) audioRef.value.playbackRate = clamped
  }

  // 切换当前在线歌曲音质（仅 QQ/酷狗生效），重建带 quality 参数的流地址并重载音频
  async function switchOnlineQuality(quality: string) {
    const om = currentOnlineSong.value
    if (!om || (om.source !== 'qq' && om.source !== 'kugou')) return

    // 解析并更新 extra 中的 quality
    let extraObj: Record<string, string> = {}
    if (om.extra) {
      try { extraObj = JSON.parse(om.extra) } catch { /* 忽略损坏的 extra */ }
    }
    extraObj['quality'] = quality
    const newExtra = JSON.stringify(extraObj)

    const updated: OnlineSong = { ...om, extra: newExtra }
    currentOnlineSong.value = updated
    onlineMap.set(onlineKey(updated), updated)

    // 重建流地址（带 quality 参数）
    const base = serverUrl.value || (await AudioServerURL())
    serverUrl.value = base
    const q = new URLSearchParams({ source: updated.source, id: updated.id })
    if (newExtra) q.set('extra', newExtra)
    q.set('quality', quality)
    updated.streamUrl = `${base}/online?${q.toString()}`

    // 同步更新队列中当前曲目路径
    if (index.value >= 0 && index.value < queue.value.length) {
      const next = [...queue.value]
      next[index.value] = { ...next[index.value], path: updated.streamUrl }
      queue.value = next
    }

    if (!audioRef.value) return
    const wasPlaying = isPlaying.value
    const resumeAt = wasPlaying ? audioRef.value.currentTime || 0 : 0

    audioRef.value.src = updated.streamUrl
    audioRef.value.load()
    audioRef.value.playbackRate = playbackRate.value

    const onMeta = () => {
      if (resumeAt > 0 && audioRef.value) {
        try { audioRef.value.currentTime = resumeAt } catch { /* 忽略 seek 失败 */ }
      }
      audioRef.value?.removeEventListener('loadedmetadata', onMeta)
    }
    audioRef.value.addEventListener('loadedmetadata', onMeta)

    if (wasPlaying) {
      try {
        await audioRef.value.play()
        isPlaying.value = true
      } catch {
        isPlaying.value = false
      }
    } else {
      isPlaying.value = false
    }
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
    queue,
    index,
    playlistId,
    hasSong,
    onlineList,
    onlineIndex,
    onlineMap,
    playSongs,
    playQueueAt,
    playOnline,
    addToQueue,
    addOnlineToQueue,
    removeFromQueue,
    clearQueue,
    togglePlay,
    pause,
    seek,
    setVolume,
    setPlaybackRate,
    switchOnlineQuality,
  }
}
