<script lang="ts" setup>
import { ref, watch, computed } from 'vue'
import { coverStyle, footerCoverVisible } from '../../composables/useSharedTransition'
import { useCoverTilt } from '../../composables/useCoverTilt'
import PlayerDetailCoverImage from './PlayerDetailCoverImage.vue'
import PlayerDetailCoverReflection from './PlayerDetailCoverReflection.vue'

const props = defineProps<{
  coverUrl: string | null
  isPlaying: boolean
  isExpanded: boolean
  showLyrics: boolean
  isAnimating: boolean
}>()

const emit = defineEmits<{
  toggleLyrics: []
}>()

const localCoverUrl = ref('')
const reflectionCoverUrl = ref('')
const detailCoverRef = ref<HTMLElement | null>(null)

const currentSongPath = computed(() => props.coverUrl || '')

watch(() => props.coverUrl, (cover) => {
  localCoverUrl.value = cover || ''
}, { immediate: true })

watch([() => props.coverUrl, () => props.isExpanded], () => {
  if (!props.isExpanded) {
    reflectionCoverUrl.value = ''
    return
  }
  const nextReflectionUrl = localCoverUrl.value || ''
  if (nextReflectionUrl !== reflectionCoverUrl.value) {
    reflectionCoverUrl.value = nextReflectionUrl
  }
}, { immediate: true })

const {
  shineX,
  shineY,
  isHovering,
  coverTransform,
  handleMouseMove,
  handleMouseEnter,
  handleMouseLeave,
} = useCoverTilt(() => props.isExpanded)

function onMouseMove(e: MouseEvent) {
  handleMouseMove(e, detailCoverRef.value)
}

function toggleLyrics() {
  emit('toggleLyrics')
}

defineExpose({ detailCoverRef })
</script>

<template>
  <div class="pointer-events-none">
    <div
      ref="detailCoverRef"
      class="cover-container"
      :class="[
        props.isExpanded ? 'expanded' : 'collapsed',
        props.showLyrics ? 'with-lyrics' : 'center',
        { animating: props.isAnimating },
      ]"
      :style="coverStyle"
      @click="toggleLyrics"
      @mousemove="onMouseMove"
      @mouseenter="handleMouseEnter"
      @mouseleave="handleMouseLeave"
    >
      <div class="cover-inner" :style="{ transform: coverTransform }">
        <PlayerDetailCoverImage
          :cover-url="localCoverUrl"
          :song-path="currentSongPath"
          :is-expanded="props.isExpanded"
        />

        <div
          v-if="props.isExpanded"
          class="shine-layer"
          :style="{
            background: `radial-gradient(circle at ${shineX}% ${shineY}%, rgba(255,255,255,0.35) 0%, transparent 50%)`,
            opacity: isHovering ? 1 : 0,
          }"
        ></div>
      </div>

      <PlayerDetailCoverReflection
        v-if="props.isExpanded"
        :reflection-url="reflectionCoverUrl"
      />
    </div>
  </div>
</template>

<style scoped>
.pointer-events-none {
  pointer-events: none;
}

.cover-container {
  position: absolute;
  aspect-ratio: 1;
  will-change: transform;
  pointer-events: auto;
  overflow: visible;
}

.cover-container.expanded {
  --cover-size: clamp(220px, 45vh, 580px);
  top: calc(45% - var(--cover-size) / 2);
  width: var(--cover-size);
  border-radius: 16px;
  box-shadow:
    0 30px 60px -12px rgba(0, 0, 0, 0.6),
    0 18px 36px -18px rgba(0, 0, 0, 0.7);
  transition:
    left 600ms cubic-bezier(0.22, 1, 0.36, 1),
    top 500ms cubic-bezier(0.22, 1, 0.36, 1),
    width 500ms cubic-bezier(0.22, 1, 0.36, 1),
    opacity 300ms ease,
    border-radius 300ms ease;
}

.cover-container.expanded.center {
  left: calc(50% - var(--cover-size) / 2);
}

.cover-container.expanded.with-lyrics {
  left: calc(28% - var(--cover-size) / 2);
}

.cover-container.expanded.animating {
  transition:
    opacity 300ms ease,
    border-radius 300ms ease;
}

.cover-container.collapsed {
  top: calc(100vh - 64px);
  left: 16px;
  width: 48px;
  border-radius: 8px;
  pointer-events: none;
  opacity: v-bind('footerCoverVisible ? 1 : 0');
}

.cover-inner {
  width: 100%;
  height: 100%;
  border-radius: inherit;
  overflow: hidden;
  position: relative;
  isolation: isolate;
  z-index: 20;
  transition: transform 240ms cubic-bezier(0.22, 1, 0.36, 1);
  transform-style: preserve-3d;
}

.shine-layer {
  position: absolute;
  inset: 0;
  z-index: 30;
  pointer-events: none;
  mix-blend-mode: overlay;
  transition: opacity 240ms ease-out;
  border-radius: inherit;
}
</style>
