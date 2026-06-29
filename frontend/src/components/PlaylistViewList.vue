<script lang="ts" setup>
import type { Song } from '../types'
import {
  displayTitle,
  displayArtist,
  displayAlbum,
  displayDuration,
} from '../composables/usePlaylistDisplay'

const props = defineProps<{
  songs: Song[]
  currentSong: Song | null
  selectedIds: Set<string>
  batchMode: boolean
}>()

const emit = defineEmits<{
  play: [song: Song]
  toggle: [id: string]
  remove: [id: string]
}>()

function isPlaying(song: Song) {
  return props.currentSong?.id === song.id
}
</script>

<template>
  <div class="song-list">
    <div class="list-header">
      <span v-if="props.batchMode" class="col-check"></span>
      <span class="col-index">#</span>
      <span class="col-title">标题</span>
      <span class="col-artist">艺术家</span>
      <span class="col-album">专辑</span>
      <span class="col-duration">时长</span>
      <span class="col-action"></span>
    </div>
    <div
      v-for="(song, index) in props.songs"
      :key="song.id"
      :class="['song-item', { playing: isPlaying(song), selected: props.selectedIds.has(song.id) }]"
      @dblclick="emit('play', song)"
    >
      <label
        v-if="props.batchMode"
        class="col-check"
        @click.stop
      >
        <input
          type="checkbox"
          :checked="props.selectedIds.has(song.id)"
          @change="emit('toggle', song.id)"
        />
      </label>
      <span class="col-index">{{ index + 1 }}</span>
      <div class="col-title">
        <div class="primary-text">{{ displayTitle(song) }}</div>
        <div class="secondary-text">{{ displayArtist(song) }}</div>
      </div>
      <span class="col-artist secondary-text">{{ displayArtist(song) }}</span>
      <span class="col-album secondary-text">{{ displayAlbum(song) }}</span>
      <span class="col-duration secondary-text">{{ displayDuration(song) }}</span>
      <button
        class="remove-btn"
        title="移除"
        @click="emit('remove', song.id)"
      >
        ✕
      </button>
    </div>
    <div v-if="props.songs.length === 0" class="empty-state">
      暂无歌曲
    </div>
  </div>
</template>

<style scoped>
.song-list {
  flex: 1;
  overflow-y: auto;
  padding: 0 16px 8px;
}

.list-header,
.song-item {
  display: grid;
  grid-template-columns: 36px 2fr 1.2fr 1.2fr 64px 28px;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
}

.list-header:has(.col-check),
.song-item:has(.col-check) {
  grid-template-columns: 32px 36px 2fr 1.2fr 1.2fr 64px 28px;
}

.list-header {
  position: sticky;
  top: 0;
  font-size: 11px;
  font-weight: 600;
  color: var(--fluent-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.4px;
  background: var(--fluent-bg-glass);
  backdrop-filter: blur(10px);
  z-index: 1;
}

.song-item {
  border-radius: 8px;
  cursor: default;
  transition: background 0.18s ease;
}

.song-item:hover {
  background: var(--fluent-bg-hover);
}

.song-item.selected {
  background: var(--fluent-bg-active);
}

.song-item.playing {
  background: var(--fluent-accent);
  color: #fff;
}

.song-item.playing .secondary-text {
  color: rgba(255, 255, 255, 0.75);
}

.song-item.playing .col-index {
  color: #fff;
}

.col-check {
  display: flex;
  align-items: center;
  justify-content: center;
}

.col-check input {
  width: 16px;
  height: 16px;
  accent-color: var(--fluent-accent);
  cursor: pointer;
}

.col-index {
  text-align: center;
  font-size: 12px;
  color: var(--fluent-text-secondary);
}

.col-title {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.primary-text {
  font-size: 13px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.secondary-text {
  font-size: 12px;
  color: var(--fluent-text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.col-artist,
.col-album,
.col-duration {
  min-width: 0;
}

.remove-btn {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: inherit;
  font-size: 11px;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.18s ease, background 0.18s ease;
}

.song-item:hover .remove-btn {
  opacity: 1;
}

.remove-btn:hover {
  background: var(--fluent-bg-active);
}

.empty-state {
  padding: 40px 20px;
  text-align: center;
  color: var(--fluent-text-secondary);
  font-size: 13px;
}
</style>
