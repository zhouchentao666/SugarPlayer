<script lang="ts" setup>
import type { OnlineSong } from '../../bindings/sugarplayer/models'
import CommentList from './CommentList.vue'

defineProps<{
  song: OnlineSong | null
  fullscreen?: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()
</script>

<template>
  <div class="player-comments" :class="{ fullscreen }">
    <div class="pc-head">
      <div class="pc-title">
        <span class="pc-label">评论</span>
        <span v-if="song" class="pc-song">{{ song.name }} - {{ song.artist }}</span>
      </div>
      <button class="pc-close" title="关闭评论" @click="emit('close')">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
          <line x1="6" y1="6" x2="18" y2="18" />
          <line x1="18" y1="6" x2="6" y2="18" />
        </svg>
      </button>
    </div>
    <div class="pc-body">
      <CommentList :song="song" />
    </div>
  </div>
</template>

<style scoped>
.player-comments {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 72px;
  height: 420px;
  max-height: 60vh;
  display: flex;
  flex-direction: column;
  background: var(--fluent-bg-glass);
  backdrop-filter: blur(40px) saturate(125%);
  -webkit-backdrop-filter: blur(40px) saturate(125%);
  border-top: 1px solid var(--fluent-border);
  box-shadow: 0 -12px 40px rgba(0, 0, 0, 0.25);
  z-index: 70;
  animation: slideUp 0.22s ease;
}

/* 全屏播放器内：评论显示在歌词区（右侧 45%），背景透明，覆盖在模糊封面上。
   底部留出 72px（播放栏高度），避免评论面板压住播放栏 */
.player-comments.fullscreen {
  position: fixed;
  left: auto;
  bottom: 72px;
  top: 0;
  right: 0;
  width: 45%;
  height: auto;
  max-height: none;
  padding: 120px 64px 32px 32px;
  background: transparent;
  backdrop-filter: none;
  -webkit-backdrop-filter: none;
  border-top: none;
  box-shadow: none;
  z-index: 80;
  animation: fadeIn 0.22s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.player-comments.fullscreen .pc-head {
  border-bottom-color: rgba(255, 255, 255, 0.14);
  padding: 0 0 10px;
}

.player-comments.fullscreen .pc-label {
  color: #fff;
}

.player-comments.fullscreen .pc-song {
  color: rgba(255, 255, 255, 0.6);
}

.player-comments.fullscreen .pc-close {
  color: rgba(255, 255, 255, 0.7);
}

.player-comments.fullscreen .pc-close:hover {
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
}

.player-comments.fullscreen .pc-body {
  padding: 0;
}

/* 透明背景下统一文字为浅色，保证可读性 */
.player-comments.fullscreen :deep(.comment-head) {
  padding: 0 0 10px;
}

.player-comments.fullscreen :deep(.seg) {
  background: rgba(255, 255, 255, 0.12);
}

.player-comments.fullscreen :deep(.seg-item) {
  color: rgba(255, 255, 255, 0.7);
}

.player-comments.fullscreen :deep(.seg-item.active) {
  background: rgba(255, 255, 255, 0.22);
  color: #fff;
}

.player-comments.fullscreen :deep(.comment-total),
.player-comments.fullscreen :deep(.c-loc),
.player-comments.fullscreen :deep(.c-time),
.player-comments.fullscreen :deep(.c-like) {
  color: rgba(255, 255, 255, 0.65);
}

.player-comments.fullscreen :deep(.c-name),
.player-comments.fullscreen :deep(.c-text) {
  color: #fff;
}

.player-comments.fullscreen :deep(.comment-item) {
  border-bottom-color: rgba(255, 255, 255, 0.12);
}

.player-comments.fullscreen :deep(.c-replies) {
  background: rgba(255, 255, 255, 0.08);
}

.player-comments.fullscreen :deep(.load-more) {
  color: #fff;
  border-color: rgba(255, 255, 255, 0.25);
}

.player-comments.fullscreen :deep(.load-more:hover:not(:disabled)) {
  background: rgba(255, 255, 255, 0.12);
}

.player-comments.fullscreen :deep(.comment-state),
.player-comments.fullscreen :deep(.comment-empty) {
  color: rgba(255, 255, 255, 0.7);
}

.player-comments.fullscreen :deep(.comment-state.error) {
  color: #ff9a9a;
}

@keyframes slideUp {
  from {
    transform: translateY(20px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

.pc-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 18px;
  border-bottom: 1px solid var(--fluent-border);
  flex-shrink: 0;
}

.pc-title {
  display: flex;
  align-items: baseline;
  gap: 10px;
  min-width: 0;
}

.pc-label {
  font-size: 14px;
  font-weight: 700;
  color: var(--fluent-text);
}

.pc-song {
  font-size: 12px;
  color: var(--fluent-text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pc-close {
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

.pc-close:hover {
  background: var(--fluent-bg-hover);
  color: var(--fluent-text);
}

.pc-close svg {
  width: 18px;
  height: 18px;
}

.pc-body {
  flex: 1;
  min-height: 0;
  padding: 0 14px;
}
</style>
