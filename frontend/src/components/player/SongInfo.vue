<script lang="ts" setup>
import { footerCoverVisible } from '../../composables/useSharedTransition'
import { type Song } from '../../types'
import { displayTitle, displayArtist } from '../../composables/usePlaylistDisplay'

const props = defineProps<{
  song: Song | null
  coverUrl: string | null
  showDetail?: boolean
}>()

const emit = defineEmits<{
  click: []
}>()

function songTitle(song: Song | null): string {
  return song ? displayTitle(song) : '未播放'
}

function songArtist(song: Song | null): string {
  return song ? displayArtist(song) : '未知艺术家'
}

function handleClick() {
  if (props.showDetail) {
    emit('click')
  }
}
</script>

<template>
  <div class="song-info-wrap">
    <div
      class="cover"
      data-footer-cover
      :class="{ hidden: !footerCoverVisible }"
      @click="emit('click')"
    >
      <img v-if="coverUrl" :src="coverUrl" class="cover-img" alt="cover" />
      <span v-else-if="!song">♪</span>
      <span v-else>♫</span>
    </div>
    <div
      class="song-info"
      :class="{ 'detail-mode': showDetail }"
      @click="handleClick"
    >
      <div class="song-title" :class="{ 'detail-mode': showDetail }">{{ songTitle(song) }}</div>
      <div class="song-artist" :class="{ 'detail-mode': showDetail }">{{ songArtist(song) }}</div>
    </div>
  </div>
</template>

<style scoped>
.song-info-wrap {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.cover {
  width: 44px;
  height: 44px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--fluent-bg-hover);
  font-size: 18px;
  color: var(--fluent-text-secondary);
  flex-shrink: 0;
  overflow: hidden;
  cursor: pointer;
  transition: transform 200ms ease, box-shadow 200ms ease;
}

.cover:hover {
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.cover.hidden {
  opacity: 0;
  pointer-events: none;
  transition: none;
}

.cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.song-info {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
  cursor: default;
  transition: transform 500ms cubic-bezier(0.22, 1, 0.36, 1);
}

.song-info.detail-mode {
  transform: translateX(-60px);
  cursor: pointer;
}

.song-title {
  font-size: 13px;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.song-artist {
  font-size: 11px;
  color: var(--fluent-text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  transition: color 500ms ease;
}

.song-artist.detail-mode {
  color: rgba(255, 255, 255, 0.7);
}
</style>
