<script lang="ts" setup>
import { ref } from 'vue'
import { type Playlist } from '../types'
import type { OnlineCollection } from '../../bindings/sugarplayer/models'
import PlaylistItem from './sidebar/PlaylistItem.vue'
import PlaylistCreateInput from './sidebar/PlaylistCreateInput.vue'

const props = defineProps<{
  playlists: Playlist[]
  selectedId: string
  activeView?: 'main' | 'settings' | 'online' | 'online-discover' | 'onlinesettings'
  pinnedCollections: OnlineCollection[]
}>()

const emit = defineEmits<{
  (e: 'update:playlists', playlists: Playlist[]): void
  (e: 'update:selectedId', id: string): void
  (e: 'open-settings'): void
  (e: 'open-search'): void
  (e: 'open-discover'): void
  (e: 'open-online-settings'): void
  (e: 'select', id: string): void
  (e: 'drop-songs', payload: { targetPlaylistId: string; sourcePlaylistId: string; songIds: string[] }): void
  (e: 'open-online-collection', collection: OnlineCollection): void
  (e: 'unpin-collection', collection: OnlineCollection): void
}>()

const sourceName: Record<string, string> = {
  netease: '网易云',
  qq: 'QQ',
  kugou: '酷狗',
  kuwo: '酷我',
  migu: '咪咕',
  bilibili: 'B站',
  soda: '汽水',
  joox: 'Joox',
  qianqian: '千千',
  apple: 'Apple',
  jamendo: 'Jamendo',
  fivesing: '5sing',
}

function openCollection(col: OnlineCollection) {
  emit('open-online-collection', col)
}

function unpin(col: OnlineCollection) {
  emit('unpin-collection', col)
}

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
    <div class="nav-section">
      <button
        :class="['nav-item', { active: activeView === 'online-discover' }]"
        @click="emit('open-discover')"
      >
        <span class="nav-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="9" />
            <path d="M3 12h18" />
            <path d="M12 3a14 14 0 0 1 0 18a14 14 0 0 1 0-18" />
          </svg>
        </span>
        <span>发现</span>
      </button>
      <button
        :class="['nav-item', { active: activeView === 'online' }]"
        @click="emit('open-search')"
      >
        <span class="nav-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="11" cy="11" r="7" />
            <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
        </span>
        <span>搜索</span>
      </button>
    </div>
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

    <div v-if="pinnedCollections.length" class="section pinned-section">
      <div class="section-title">在线歌单</div>
      <ul class="playlist-list">
        <li
          v-for="col in pinnedCollections"
          :key="col.source + ':' + col.kind + ':' + col.id"
          class="pinned-item"
          @click="openCollection(col)"
        >
          <span class="pinned-cover">
            <img v-if="col.cover" :src="col.cover" alt="" loading="lazy" />
            <span v-else class="pinned-cover-fallback">♪</span>
          </span>
          <span class="pinned-meta">
            <span class="pinned-name">{{ col.name }}</span>
            <span class="pinned-sub">{{ sourceName[col.source] || col.source }}{{ col.kind === 'album' ? ' · 专辑' : '' }}</span>
          </span>
          <button
            class="pinned-unpin"
            title="取消固定"
            @click.stop="unpin(col)"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </li>
      </ul>
    </div>
    <div class="bottom">
      <button
        :class="['settings-btn', { active: activeView === 'onlinesettings' }]"
        @click="emit('open-online-settings')"
      >
        <span class="icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="8" r="4" />
            <path d="M4 21a8 8 0 0 1 16 0" />
          </svg>
        </span>
        <span>在线设置</span>
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
</style>