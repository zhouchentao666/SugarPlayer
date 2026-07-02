<script lang="ts" setup>
import type { Song } from '../../types'
import { displayTitle, displayArtist } from '../../composables/usePlaylistDisplay'

const props = defineProps<{
  isVisible: boolean
  isTopChromeVisible: boolean
  isMaximised: boolean
  currentSong: Song | null
  staggerStyle: (phase: number, translateDir?: 'Y' | 'X', distance?: number) => Record<string, string | number>
}>()

const emit = defineEmits<{
  close: []
  minimize: []
  toggleMaximize: []
  closeApp: []
  showTopChrome: []
  topChromeLeave: []
}>()

const handleClose = () => emit('close')
const minimize = () => emit('minimize')
const toggleMaximize = () => emit('toggleMaximize')
const closeApp = () => emit('closeApp')
const showTopChrome = () => emit('showTopChrome')
const topChromeLeave = () => emit('topChromeLeave')
</script>

<template>
  <div
    class="top-chrome"
    :style="staggerStyle(1, 'Y', -10)"
    @mouseenter="showTopChrome"
    @mousemove="showTopChrome"
    @mouseleave="topChromeLeave"
  >
    <div
      class="chrome-hitbox"
      :class="{ 'pe-auto': isVisible, 'pe-none': !isVisible }"
    ></div>

    <div
      class="chrome-inner"
      :class="{
        'translate-y-0 opacity-100': isTopChromeVisible,
        '-translate-y-3 opacity-0': !isTopChromeVisible,
        'pe-auto': isVisible,
        'pe-none': !isVisible,
      }"
    >
      <div class="drag-region" style="--wails-draggable: drag"></div>

      <div class="chrome-left">
        <button title="收起详情页" class="chrome-btn" @click="handleClose">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
          </svg>
        </button>
      </div>

      <div class="chrome-center">
        {{ currentSong ? displayTitle(currentSong) : '' }}
        <span v-if="currentSong && displayArtist(currentSong) !== '未知艺术家'" class="mx-1 opacity-60">-</span>
        <span class="opacity-60">{{ currentSong ? displayArtist(currentSong) : '' }}</span>
      </div>

      <div class="chrome-right">
        <button class="chrome-btn small" title="最小化" @click="minimize">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M5 12h14" />
          </svg>
        </button>
        <button class="chrome-btn small" :title="props.isMaximised ? '还原' : '最大化'" @click="toggleMaximize">
          <svg v-if="props.isMaximised" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="6" y="6" width="14" height="14" rx="2" ry="2" />
            <rect x="3" y="3" width="14" height="14" rx="2" ry="2" />
          </svg>
          <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
          </svg>
        </button>
        <button class="chrome-btn close-btn" title="关闭" @click="closeApp">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M18 6L6 18M6 6l12 12" />
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.top-chrome {
  position: relative;
  z-index: 60;
  height: 96px;
}

.chrome-hitbox {
  position: absolute;
  left: 0;
  right: 0;
  top: 0;
  height: 96px;
}

.chrome-inner {
  position: absolute;
  left: 0;
  right: 0;
  top: 0;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  transition: all 500ms ease-out;
}

.drag-region {
  position: absolute;
  inset: 0;
}

.chrome-left {
  position: relative;
  z-index: 10;
  display: flex;
  width: 25%;
  align-items: center;
}

.chrome-center {
  pointer-events: none;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  padding: 0 16px;
  text-align: center;
  font-size: 14px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.8);
  text-shadow: 0 2px 8px rgba(0, 0, 0, 0.4);
}

.chrome-right {
  position: relative;
  z-index: 10;
  display: flex;
  width: 25%;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
}

.chrome-btn {
  padding: 8px;
  border-radius: 8px;
  color: rgba(255, 255, 255, 0.5);
  transition: all 200ms;
  background: transparent;
  border: none;
  cursor: pointer;
}

.chrome-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: white;
}

.chrome-btn.close-btn:hover {
  background: #ef4444;
  color: white;
}

.pe-auto {
  pointer-events: auto;
}

.pe-none {
  pointer-events: none;
}

.translate-y-0 {
  transform: translateY(0);
}

.-translate-y-3 {
  transform: translateY(-12px);
}

.opacity-100 {
  opacity: 1;
}

.opacity-0 {
  opacity: 0;
}

.opacity-60 {
  opacity: 0.6;
}

.h-6 { height: 24px; }
.w-6 { width: 24px; }
.h-4 { height: 16px; }
.w-4 { width: 16px; }
.h-3 { height: 12px; }
.w-3 { width: 12px; }

.mx-1 {
  margin-left: 4px;
  margin-right: 4px;
}
</style>
