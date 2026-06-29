<script lang="ts" setup>
const props = defineProps<{
  coverUrl: string
  songPath: string
  isExpanded: boolean
}>()
</script>

<template>
  <img
    v-if="coverUrl"
    :key="`thumb:${songPath}:${coverUrl}`"
    :src="coverUrl"
    class="cover-img"
    :class="isExpanded ? 'expanded-img' : 'collapsed-img'"
    draggable="false"
    decoding="async"
  />
  <div v-else class="cover-placeholder">
    <svg xmlns="http://www.w3.org/2000/svg" class="h-32 w-32" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3" />
    </svg>
  </div>
</template>

<style scoped>
.cover-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  user-select: none;
  transition: transform 240ms ease-out, filter 240ms ease-out, opacity 240ms ease-out;
  z-index: 10;
}

.expanded-img {
  transform: scale(1);
  filter: blur(0);
}

.collapsed-img {
  transform: scale(1.25);
  filter: blur(0);
}

.cover-placeholder {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  background: rgba(255, 255, 255, 0.05);
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgba(255, 255, 255, 0.1);
  z-index: 0;
}

.h-32 { height: 128px; }
.w-32 { width: 128px; }
</style>
