<script lang="ts" setup>
import { ref } from 'vue'
import type { Playlist } from '../types'

const props = defineProps<{
  show: boolean
  playlists: Playlist[]
  songName: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'select', playlistId: string): void
  (e: 'create', name: string): void
}>()

const newName = ref('')

function submitCreate() {
  const name = newName.value.trim() || `歌单 ${props.playlists.length + 1}`
  emit('create', name)
  newName.value = ''
}
</script>

<template>
  <Transition name="modal-fade">
    <div v-if="show" class="add-modal-mask" @click.self="emit('close')">
      <div class="add-modal">
        <div class="am-head">
          <div class="am-title">
            <span>添加到歌单</span>
            <span v-if="songName" class="am-song">{{ songName }}</span>
          </div>
          <button class="am-close" title="关闭" @click="emit('close')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
              <line x1="6" y1="6" x2="18" y2="18" />
              <line x1="18" y1="6" x2="6" y2="18" />
            </svg>
          </button>
        </div>

        <div class="am-list">
          <button
            v-for="pl in playlists"
            :key="pl.id"
            class="am-item"
            @click="emit('select', pl.id)"
          >
            <span class="am-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round">
                <path d="M9 18V5l12-2v13" />
                <circle cx="6" cy="18" r="3" />
                <circle cx="18" cy="16" r="3" />
              </svg>
            </span>
            <span class="am-name">{{ pl.name }}</span>
            <span class="am-count">{{ pl.songs.length }}</span>
          </button>
          <div v-if="playlists.length === 0" class="am-empty">还没有歌单</div>
        </div>

        <div class="am-create">
          <input
            v-model="newName"
            class="am-input"
            type="text"
            placeholder="新建歌单名称"
            @keydown.enter="submitCreate"
          />
          <button class="am-create-btn" @click="submitCreate">新建并添加</button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.add-modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 140;
}

.add-modal {
  width: min(420px, 92vw);
  max-height: 76vh;
  display: flex;
  flex-direction: column;
  background: var(--fluent-bg-glass);
  backdrop-filter: blur(40px) saturate(125%);
  -webkit-backdrop-filter: blur(40px) saturate(125%);
  border: 1px solid var(--fluent-border);
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.4);
  overflow: hidden;
}

.am-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px;
  border-bottom: 1px solid var(--fluent-border);
  flex-shrink: 0;
}

.am-title {
  display: flex;
  align-items: baseline;
  gap: 10px;
  min-width: 0;
}

.am-title > span:first-child {
  font-size: 15px;
  font-weight: 700;
  color: var(--fluent-text);
}

.am-song {
  font-size: 12px;
  color: var(--fluent-text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.am-close {
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 50%;
  background: transparent;
  color: var(--fluent-text-secondary);
  cursor: pointer;
  flex-shrink: 0;
  transition: background 0.18s ease, color 0.18s ease;
}

.am-close:hover {
  background: var(--fluent-bg-hover);
  color: var(--fluent-text);
}

.am-close svg {
  width: 18px;
  height: 18px;
}

.am-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
  min-height: 0;
}

.am-item {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 10px 12px;
  border: none;
  border-radius: 10px;
  background: transparent;
  color: var(--fluent-text);
  font-size: 13px;
  text-align: left;
  cursor: pointer;
  transition: background 0.18s ease;
}

.am-item:hover {
  background: var(--fluent-bg-hover);
}

.am-icon {
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  background: var(--fluent-bg-active);
  color: var(--fluent-text-secondary);
  flex-shrink: 0;
}

.am-icon svg {
  width: 18px;
  height: 18px;
}

.am-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 500;
}

.am-count {
  font-size: 12px;
  color: var(--fluent-text-secondary);
  flex-shrink: 0;
}

.am-empty {
  padding: 30px 12px;
  text-align: center;
  color: var(--fluent-text-secondary);
  font-size: 13px;
}

.am-create {
  display: flex;
  gap: 8px;
  padding: 12px;
  border-top: 1px solid var(--fluent-border);
  flex-shrink: 0;
}

.am-input {
  flex: 1;
  height: 36px;
  padding: 0 12px;
  border: 1px solid var(--fluent-input-border);
  border-radius: 10px;
  background: var(--fluent-input-bg);
  color: var(--fluent-text);
  font-size: 13px;
  outline: none;
}

.am-input:focus {
  border-color: var(--fluent-accent);
}

.am-create-btn {
  flex-shrink: 0;
  height: 36px;
  padding: 0 16px;
  border: none;
  border-radius: 10px;
  background: var(--fluent-accent);
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: filter 0.18s ease;
}

.am-create-btn:hover {
  filter: brightness(1.1);
}

.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.2s ease;
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}
</style>
