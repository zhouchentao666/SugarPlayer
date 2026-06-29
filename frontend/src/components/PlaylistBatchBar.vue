<script lang="ts" setup>
import { ref, computed } from 'vue'
import type { Playlist, Song } from '../types'

const props = defineProps<{
  selectedSongs: Song[]
  playlists: Playlist[]
  currentPlaylistId: string
}>()

const emit = defineEmits<{
  remove: []
  'add-to-playlist': [playlistId: string]
  'replace-to-playlist': [playlistId: string]
  'select-all': []
  'clear-selection': []
  close: []
}>()

const targetId = ref('')

const targetPlaylists = computed(() =>
  props.playlists.filter(p => p.id !== props.currentPlaylistId)
)

const hasTarget = computed(() => targetPlaylists.value.some(p => p.id === targetId.value))

function handleAdd() {
  if (!hasTarget.value) return
  emit('add-to-playlist', targetId.value)
}

function handleReplace() {
  if (!hasTarget.value) return
  emit('replace-to-playlist', targetId.value)
}
</script>

<template>
  <div class="batch-bar" style="--wails-draggable: no-drag;">
    <div class="batch-info">
      已选 <strong>{{ props.selectedSongs.length }}</strong> 首
    </div>
    <div class="batch-actions">
      <button class="batch-btn" @click="emit('select-all')">
        全选
      </button>
      <button class="batch-btn" @click="emit('clear-selection')">
        清空
      </button>
      <button class="batch-btn danger" @click="emit('remove')">
        从歌单移除
      </button>
      <select
        v-model="targetId"
        class="batch-select"
      >
        <option value="" disabled>选择目标歌单</option>
        <option
          v-for="playlist in targetPlaylists"
          :key="playlist.id"
          :value="playlist.id"
        >
          {{ playlist.name }}
        </option>
      </select>
      <button
        class="batch-btn"
        :disabled="!hasTarget"
        @click="handleAdd"
      >
        添加到歌单
      </button>
      <button
        class="batch-btn"
        :disabled="!hasTarget"
        @click="handleReplace"
      >
        替换到歌单
      </button>
      <button class="batch-btn ghost" @click="emit('close')">
        完成
      </button>
    </div>
  </div>
</template>

<style scoped>
.batch-bar {
  position: absolute;
  left: 50%;
  bottom: 20px;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 10px 16px;
  border-radius: 12px;
  background: var(--fluent-bg-glass);
  backdrop-filter: blur(20px) saturate(150%);
  -webkit-backdrop-filter: blur(20px) saturate(150%);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.25);
  border: 1px solid var(--fluent-border);
  z-index: 20;
  animation: slide-up 0.25s cubic-bezier(0.22, 1, 0.36, 1);
}

@keyframes slide-up {
  from {
    opacity: 0;
    transform: translate(-50%, 20px);
  }
  to {
    opacity: 1;
    transform: translate(-50%, 0);
  }
}

.batch-info {
  font-size: 13px;
  white-space: nowrap;
  color: var(--fluent-text);
}

.batch-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.batch-select {
  padding: 7px 10px;
  border: 1px solid var(--fluent-border);
  border-radius: 8px;
  background: var(--fluent-bg-card);
  color: var(--fluent-text);
  font-size: 13px;
  outline: none;
  cursor: pointer;
  max-width: 140px;
}

.batch-btn {
  padding: 7px 12px;
  border: 1px solid var(--fluent-border);
  border-radius: 8px;
  background: var(--fluent-bg-card);
  color: var(--fluent-text);
  font-size: 13px;
  cursor: pointer;
  transition: background 0.18s ease, opacity 0.18s ease;
}

.batch-btn:hover:not(:disabled) {
  background: var(--fluent-bg-hover);
}

.batch-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.batch-btn.danger {
  border-color: rgba(255, 70, 70, 0.4);
  color: rgba(255, 90, 90, 1);
}

.batch-btn.danger:hover:not(:disabled) {
  background: rgba(255, 70, 70, 0.12);
}

.batch-btn.ghost {
  background: transparent;
}
</style>
