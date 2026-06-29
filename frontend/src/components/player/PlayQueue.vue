<script lang="ts" setup>
import { computed } from 'vue'
import type { Playlist, Song } from '../../types'

const props = defineProps<{
  show: boolean
  playlist: Playlist | null
  currentSong: Song | null
}>()

const emit = defineEmits<{
  close: []
  play: [index: number]
}>()

const songs = computed(() => props.playlist?.songs || [])

function formatDuration(seconds: number): string {
  if (!seconds || seconds < 0) return '0:00'
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

function isActive(song: Song): boolean {
  return props.currentSong?.id === song.id
}
</script>

<template>
  <div
    class="queue-panel"
    :class="{ visible: show }"
  >
    <div class="queue-content">
      <div class="queue-header">
        <span class="queue-title">播放列表</span>
        <button class="queue-close" @click="emit('close')">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M18 6L6 18M6 6l12 12" />
          </svg>
        </button>
      </div>
      <div class="queue-body">
        <div
          v-for="(song, index) in songs"
          :key="song.id"
          class="queue-item"
          :class="{ active: isActive(song) }"
          @click="emit('play', index)"
        >
          <span class="item-index">{{ index + 1 }}</span>
          <div class="item-info">
            <div class="item-title">{{ song.metadata?.title || song.title }}</div>
            <div class="item-artist">{{ song.metadata?.artist || '未知艺术家' }}</div>
          </div>
          <span class="item-duration">{{ formatDuration(song.metadata?.duration || 0) }}</span>
        </div>
        <div v-if="songs.length === 0" class="queue-empty">暂无歌曲</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.queue-panel {
  position: fixed;
  right: 12px;
  bottom: 84px;
  z-index: 70;
  width: 320px;
  max-height: calc(100vh - 96px);
  display: flex;
  flex-direction: column;
  background: var(--fluent-bg-glass);
  backdrop-filter: blur(40px) saturate(125%);
  -webkit-backdrop-filter: blur(40px) saturate(125%);
  border: 1px solid var(--fluent-border);
  border-radius: 8px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.25);
  color: var(--fluent-text);
  opacity: 0;
  pointer-events: none;
  visibility: hidden;
  transform: translateY(10px) scale(0.98);
  transform-origin: bottom right;
  transition: opacity 250ms ease, transform 250ms cubic-bezier(0.22, 1, 0.36, 1), visibility 250ms;
  overflow: hidden;
}

.queue-panel.visible {
  opacity: 1;
  pointer-events: auto;
  visibility: visible;
  transform: translateY(0) scale(1);
}

.queue-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.queue-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid var(--fluent-border);
  flex-shrink: 0;
}

.queue-title {
  font-size: 15px;
  font-weight: 600;
}

.queue-close {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 50%;
  background: transparent;
  color: var(--fluent-text-secondary);
  cursor: pointer;
  transition: background 0.18s ease, color 0.18s ease;
}

.queue-close:hover {
  background: var(--fluent-bg-hover);
  color: var(--fluent-text);
}

.queue-close svg {
  width: 16px;
  height: 16px;
}

.queue-body {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}

.queue-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  cursor: pointer;
  transition: background 0.15s ease;
}

.queue-item:hover {
  background: var(--fluent-bg-hover);
}

.queue-item.active {
  background: var(--fluent-bg-active);
  color: var(--fluent-accent);
}

.item-index {
  width: 24px;
  text-align: center;
  font-size: 12px;
  color: var(--fluent-text-secondary);
  flex-shrink: 0;
}

.queue-item.active .item-index {
  color: var(--fluent-accent);
}

.item-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.item-title {
  font-size: 13px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-artist {
  font-size: 11px;
  color: var(--fluent-text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-duration {
  font-size: 11px;
  color: var(--fluent-text-secondary);
  flex-shrink: 0;
}

.queue-empty {
  padding: 32px 16px;
  text-align: center;
  font-size: 13px;
  color: var(--fluent-text-secondary);
}
</style>
