<script lang="ts" setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { Events } from '@wailsio/runtime'
import TitleBar from './components/TitleBar.vue'
import Sidebar from './components/Sidebar.vue'
import Settings from './components/Settings.vue'
import PlaylistView from './components/PlaylistView.vue'
import PlayerFooter from './components/PlayerFooter.vue'
import PlayerDetail from './components/player/PlayerDetail.vue'
import { useAudioPlayer } from './composables/useAudioPlayer'
import { usePlaylists } from './composables/usePlaylists'
import { useConfig, type AppSettings, type ConfigPlayback, type ConfigWindow } from './composables/useConfig'
import { useLyrics } from './composables/useLyrics'
import { useWindowEffect } from './composables/useWindowEffect'
import { useSession } from './composables/useSession'
import PlayQueue from './components/player/PlayQueue.vue'
import SongEditor from './components/SongEditor.vue'
import type { PlayMode } from './components/player/PlayerControls.vue'
import type { Song } from './types'
import type { SortMode, SortOrder } from './composables/usePlaylistView'
import { localMetadata } from './composables/useLocalMetadata'

const view = ref<'main' | 'settings'>('main')
const isLoading = ref(true)
const audioRef = ref<HTMLAudioElement | null>(null)
const showPlayerDetail = ref(false)
const showQueue = ref(false)
const showSongEditor = ref(false)
const editingSong = ref<Song | null>(null)
const playMode = ref<PlayMode>('sequential')
const settings = ref<AppSettings>({
  theme: 'system',
  accentColor: '#0078d4',
  quality: 'standard',
  autoplay: false,
  savePlaylistAndSong: true,
  saveWindowPosition: true,
  windowEffect: 'acrylic',
  customImagePath: '',
  customImageOpacity: 35,
  customImageBlur: 20,
  songColorOpacity: 45,
  songColorBlur: 30,
  fullScreenBackground: 'static',
  immersivePlayerBar: false,
  selectedPlaylistId: '',
  playlistSorts: {},
  localMetadata: {},
})

const playbackState = ref<ConfigPlayback>({
  playlistId: '',
  songIndex: -1,
  time: 0,
})

const windowState = ref<ConfigWindow>({
  x: 0,
  y: 0,
  width: 800,
  height: 600,
})

const { playlists, selectedId, updatePlaylists, updatePlaylist, selectPlaylist, addMusicFiles, addMusicFolder, refreshFolder, rewatchFolders, addSongs, replaceSongs } = usePlaylists()
const currentPlaylist = computed(() => playlists.value.find(p => p.id === selectedId.value))
const currentPlaylistSort = computed(() => currentPlaylist.value ? settings.value.playlistSorts[currentPlaylist.value.id] : undefined)

const { save, load } = useConfig(playlists, settings, playbackState, windowState, isLoading)

const audio = useAudioPlayer({
  audioRef,
  onEnded: playNext,
})

const lyrics = useLyrics(audio.currentSong)
const { appStyle, layerStyle } = useWindowEffect(settings, audio.coverUrl)

const lyricTime = computed(() => Math.floor(audio.currentTime.value * 1000))

function pickRandomIndex(current: number, count: number): number {
  if (count <= 1) return 0
  let nextIndex = current
  do { nextIndex = Math.floor(Math.random() * count) } while (nextIndex === current)
  return nextIndex
}

function playNext() {
  if (!audio.playlistId.value) return
  const playlist = playlists.value.find(p => p.id === audio.playlistId.value)
  if (!playlist || playlist.songs.length === 0) return
  const count = playlist.songs.length
  const current = audio.index.value

  if (playMode.value === 'stop') return

  let nextIndex = current
  if (playMode.value === 'shuffle') {
    nextIndex = pickRandomIndex(current, count)
  } else if (playMode.value === 'single') {
    nextIndex = current
  } else if (playMode.value === 'reverse') {
    nextIndex = current - 1
    if (nextIndex < 0) nextIndex = count - 1
  } else {
    nextIndex = current + 1
    if (nextIndex >= count) nextIndex = 0
  }
  playSong(audio.playlistId.value, nextIndex)
}

function playPrev() {
  if (!audio.playlistId.value) return
  const playlist = playlists.value.find(p => p.id === audio.playlistId.value)
  if (!playlist || playlist.songs.length === 0) return
  const count = playlist.songs.length
  const current = audio.index.value

  let prevIndex = current
  if (playMode.value === 'shuffle') {
    prevIndex = pickRandomIndex(current, count)
  } else if (playMode.value === 'single') {
    prevIndex = current
  } else if (playMode.value === 'reverse') {
    prevIndex = current + 1
    if (prevIndex >= count) prevIndex = 0
  } else {
    prevIndex = current - 1
    if (prevIndex < 0) prevIndex = count - 1
  }
  playSong(audio.playlistId.value, prevIndex)
}

function playSong(playlistId: string, index: number, autoPlay = true) {
  const playlist = playlists.value.find(p => p.id === playlistId)
  if (!playlist || index < 0 || index >= playlist.songs.length) return
  audio.play(playlistId, index, playlist.songs[index], autoPlay)
}

function playCurrentSong(index: number) {
  if (!currentPlaylist.value) return
  playSong(currentPlaylist.value.id, index)
}

function handleTogglePlay() {
  if (!audio.currentSong.value && currentPlaylist.value?.songs.length) {
    playSong(currentPlaylist.value.id, 0)
    return
  }
  audio.togglePlay()
}

function updateSettings(newSettings: AppSettings) {
  settings.value = { ...newSettings }
}
const { handleClose, restoreSession } = useSession(settings, playbackState, windowState, save, playlists, audio, selectPlaylist)

function onSelectPlaylist(id: string) {
  selectPlaylist(id)
  settings.value.selectedPlaylistId = id
  view.value = 'main'
}

watch(selectedId, (id) => {
  if (id) settings.value.selectedPlaylistId = id
})

function handleDropSongs(payload: { targetPlaylistId: string; sourcePlaylistId: string; songIds: string[] }) {
  const source = playlists.value.find(p => p.id === payload.sourcePlaylistId)
  if (!source) return
  const songs = payload.songIds
    .map(id => source.songs.find(s => s.id === id))
    .filter((s): s is Song => Boolean(s))
  if (songs.length) addSongs(payload.targetPlaylistId, songs)
}

function handleUpdateSort(payload: { playlistId: string; mode: SortMode; order: SortOrder }) {
  const next = { ...settings.value.playlistSorts }
  next[payload.playlistId] = { mode: payload.mode, order: payload.order }
  settings.value.playlistSorts = next
}

function togglePlayerDetail() {
  showPlayerDetail.value = !showPlayerDetail.value
}

function cyclePlayMode() {
  const modes: PlayMode[] = ['sequential', 'single', 'reverse', 'stop', 'shuffle']
  playMode.value = modes[(modes.indexOf(playMode.value) + 1) % modes.length]
}

function toggleQueue() {
  showQueue.value = !showQueue.value
}

function handleEditSong(song: Song) {
  editingSong.value = song
  showSongEditor.value = true
}

function closeSongEditor() {
  showSongEditor.value = false
  editingSong.value = null
}

let offFolderChanged: (() => void) | null = null

onMounted(async () => {
  await load()
  localMetadata.value = settings.value.localMetadata
  if (settings.value.selectedPlaylistId && playlists.value.some(p => p.id === settings.value.selectedPlaylistId)) {
    selectPlaylist(settings.value.selectedPlaylistId)
  }
  await rewatchFolders()
  await restoreSession()
  offFolderChanged = Events.On('folder:changed', (event: any) => {
    refreshFolder(event.data)
  })
})

onUnmounted(() => {
  offFolderChanged?.()
  Events.Off('folder:changed')
})
</script>

<template>
  <div
    class="glass"
    :data-theme="settings.theme"
    :style="appStyle"
  >
    <div
      v-if="layerStyle"
      class="window-bg-layer"
      :style="layerStyle"
    ></div>
    <TitleBar @close="handleClose" />
    <div class="content">
      <Sidebar
        :playlists="playlists"
        :selected-id="selectedId"
        :active-view="view"
        @update:playlists="updatePlaylists"
        @update:selected-id="selectedId = $event"
        @open-settings="view = 'settings'"
        @select="onSelectPlaylist"
        @drop-songs="handleDropSongs"
      />
      <main class="main">
        <PlaylistView
          v-if="view === 'main' && currentPlaylist"
          :playlist="currentPlaylist"
          :playlists="playlists"
          :current-song="audio.currentSong.value"
          :initial-sort="currentPlaylistSort"
          @update:playlist="updatePlaylist"
          @add-music-files="currentPlaylist && addMusicFiles(currentPlaylist.id)"
          @add-music-folder="currentPlaylist && addMusicFolder(currentPlaylist.id)"
          @play-song="playCurrentSong"
          @add-to-playlist="addSongs"
          @replace-to-playlist="replaceSongs"
          @update-sort="handleUpdateSort"
          @edit="handleEditSong"
        />
        <Settings
          v-if="view === 'settings'"
          :settings="settings"
          @update:settings="updateSettings"
          @close="view = 'main'"
        />
      </main>
    </div>
    <PlayerFooter
      :current-song="audio.currentSong.value"
      :cover-url="audio.coverUrl.value"
      :is-playing="audio.isPlaying.value"
      :current-time="audio.currentTime.value"
      :duration="audio.duration.value"
      :volume="audio.volume.value"
      :playback-rate="audio.playbackRate.value"
      :show-detail="showPlayerDetail"
      :play-mode="playMode"
      :immersive="settings.immersivePlayerBar"
      @toggle-play="handleTogglePlay"
      @prev="playPrev"
      @next="playNext"
      @seek="audio.seek"
      @set-volume="audio.setVolume"
      @set-playback-rate="audio.setPlaybackRate"
      @open-detail="togglePlayerDetail"
      @cycle-mode="cyclePlayMode"
      @toggle-queue="toggleQueue"
    />
    <PlayerDetail
      :show="showPlayerDetail"
      :current-song="audio.currentSong.value"
      :cover-url="audio.coverUrl.value"
      :is-playing="audio.isPlaying.value"
      :lyrics="lyrics.lyrics.value"
      :has-lyrics="lyrics.hasLyrics.value"
      :current-time="lyricTime"
      :background-mode="settings.fullScreenBackground"
      :immersive-player-bar="settings.immersivePlayerBar"
      @close="togglePlayerDetail"
      @seek="audio.seek"
    />
    <PlayQueue
      :show="showQueue"
      :playlist="currentPlaylist ?? null"
      :current-song="audio.currentSong.value"
      @close="showQueue = false"
      @play="playCurrentSong"
    />
    <SongEditor
      v-if="showSongEditor && editingSong"
      :song="editingSong"
      @close="closeSongEditor"
    />
    <audio ref="audioRef" style="display: none;"></audio>
  </div>
</template>

<style scoped>
.glass {
  position: relative;
  width: 100vw;
  height: 100vh;
  display: flex;
  flex-direction: column;
  color: var(--fluent-text);
  background: var(--fluent-bg-glass);
  backdrop-filter: blur(40px) saturate(125%);
  -webkit-backdrop-filter: blur(40px) saturate(125%);
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.1);
}

.window-bg-layer {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: -1;
}

.content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.main {
  flex: 1;
  overflow: hidden;
}
</style>
