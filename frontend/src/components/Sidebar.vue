<script lang="ts" setup>
import { ref } from 'vue'
import { type Playlist } from '../types'
import PlaylistItem from './sidebar/PlaylistItem.vue'
import PlaylistCreateInput from './sidebar/PlaylistCreateInput.vue'

const props = defineProps<{
  playlists: Playlist[]
  selectedId: string
  activeView?: 'main' | 'settings' | 'donate'
}>()

const emit = defineEmits<{
  (e: 'update:playlists', playlists: Playlist[]): void
  (e: 'update:selectedId', id: string): void
  (e: 'open-settings'): void
  (e: 'open-donate'): void
  (e: 'select', id: string): void
  (e: 'drop-songs', payload: { targetPlaylistId: string; sourcePlaylistId: string; songIds: string[] }): void
}>()

const isCreating = ref(false)

function updatePlaylists(updated: Playlist[]) {
  emit('update:playlists', updated)
}

function updatePlaylist(updated: Playlist) {
  updatePlaylists(props.playlists.map(p => (p.id === updated.id ? updated : p)))
}

function onSelect(id: string) {
  emit('update:selectedId', id)
  emit('select', id)
}

function onRename(id: string, name: string) {
  const playlist = props.playlists.find(p => p.id === id)
  if (playlist) updatePlaylist({ ...playlist, name })
}

function onDelete(id: string) {
  if (id === 'favorites') return
  const filtered = props.playlists.filter(p => p.id !== id)
  updatePlaylists(filtered)
  if (props.selectedId === id) {
    const nextId = filtered[0]?.id || ''
    emit('update:selectedId', nextId)
    emit('select', nextId)
  }
}

function startCreate() {
  isCreating.value = true
}

function confirmCreate(name: string) {
  isCreating.value = false
  if (!name) return
  const playlist: Playlist = {
    id: Date.now().toString(),
    name,
    songs: [],
    folders: [],
  }
  updatePlaylists([...props.playlists, playlist])
  onSelect(playlist.id)
}

function cancelCreate() {
  isCreating.value = false
}

function onDropSongs(playlistId: string, payload: { sourcePlaylistId: string; songIds: string[] }) {
  emit('drop-songs', { targetPlaylistId: playlistId, ...payload })
}
</script>

<template>
  <aside class="sidebar">
    <div class="section">
      <div class="section-title">歌单</div>
      <ul class="playlist-list">
        <PlaylistItem
          v-for="playlist in playlists"
          :key="playlist.id"
          :playlist="playlist"
          :selected="selectedId === playlist.id"
          @select="onSelect"
          @rename="onRename"
          @delete="onDelete"
          @drop-songs="payload => onDropSongs(playlist.id, payload)"
        />
      </ul>
      <PlaylistCreateInput
        v-if="isCreating"
        @confirm="confirmCreate"
        @cancel="cancelCreate"
      />
      <button v-else class="create-btn" @click="startCreate">
        <span class="icon">+</span>
        <span>新建歌单</span>
      </button>
    </div>

    <div class="bottom">
      <button
        class="settings-btn donate-btn"
        :class="{ active: activeView === 'donate' }"
        @click="emit('open-donate')"
      >
        <span class="icon">♥</span>
        <span>赞助作者</span>
      </button>
      <button
        :class="['settings-btn', { active: activeView === 'settings' }]"
        @click="emit('open-settings')"
      >
        <span class="icon">⚙</span>
        <span>本地设置</span>
      </button>
    </div>
  </aside>
</template>

<style scoped>
.sidebar {
  width: 220px;
  height: 100%;
  display: flex;
  flex-direction: column;
  color: var(--fluent-text);
  background: var(--fluent-bg-sidebar);
  border-right: 1px solid var(--fluent-border);
  user-select: none;
}

.section {
  flex: 1;
  padding: 16px 12px;
  overflow-y: auto;
}

.nav-section {
  padding: 12px 12px 4px;
}

.nav-item {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 12px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: inherit;
  font-size: 13px;
  cursor: pointer;
  transition: background 0.18s ease;
}

.nav-item:hover {
  background: var(--fluent-bg-hover);
}

.nav-item.active {
  background: var(--fluent-bg-active);
}

.nav-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  opacity: 0.85;
}

.nav-icon svg {
  width: 18px;
  height: 18px;
}

.section-title {
  font-size: 12px;
  font-weight: 600;
  padding: 0 10px 10px;
  color: var(--fluent-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.6px;
}

.playlist-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.create-btn,
.settings-btn {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 12px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: inherit;
  font-size: 13px;
  cursor: pointer;
  transition: background 0.18s ease;
}

.create-btn:hover,
.settings-btn:hover {
  background: var(--fluent-bg-hover);
}

.settings-btn.active {
  background: var(--fluent-bg-active);
}

.create-btn .icon,
.settings-btn .icon {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  opacity: 0.85;
}

.settings-btn .icon svg {
  width: 16px;
  height: 16px;
}

/* 在线歌单固定区 */
.pinned-section {
  flex: 0 0 auto;
}

.pinned-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 13px;
  transition: background 0.18s ease;
}

.pinned-item:hover {
  background: var(--fluent-bg-hover);
}

.pinned-cover {
  width: 16px;
  height: 16px;
  border-radius: 4px;
  overflow: hidden;
  background: var(--fluent-bg-active);
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.pinned-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.pinned-cover-fallback {
  font-size: 12px;
  color: var(--fluent-text-secondary);
}

.pinned-meta {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.pinned-name {
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pinned-sub {
  font-size: 11px;
  color: var(--fluent-text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pinned-unpin {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 50%;
  background: transparent;
  color: var(--fluent-text-secondary);
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.18s ease, background 0.18s ease, color 0.18s ease;
  flex-shrink: 0;
}

.pinned-item:hover .pinned-unpin {
  opacity: 1;
}

.pinned-unpin:hover {
  background: var(--fluent-bg-active);
  color: #ff8080;
}

.pinned-unpin svg {
  width: 14px;
  height: 14px;
}

.bottom {
  padding: 10px 12px;
  border-top: 1px solid var(--fluent-border);
}

/* 赞助作者 */
.donate-btn {
  color: #ff7a90;
}

.donate-btn:hover {
  background: rgba(255, 122, 144, 0.14);
}

.donate-btn.active {
  background: rgba(255, 122, 144, 0.18);
}

.donate-btn .icon {
  color: #ff7a90;
}
</style>