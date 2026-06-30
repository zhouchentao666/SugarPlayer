<script lang="ts" setup>
import { computed, ref, watch, nextTick, onMounted } from 'vue'
import PlayerDetailBackground from './PlayerDetailBackground.vue'
import PlayerDetailLeft from './PlayerDetailLeft.vue'
import PlayerDetailTopChrome from './PlayerDetailTopChrome.vue'
import PlayerDetailLyrics from './PlayerDetailLyrics.vue'
import { usePlayerDetail } from '../../composables/usePlayerDetail'
import type { Song } from '../../types'
import type { LyricLine } from '../../composables/useLyrics'

const props = defineProps<{
  show: boolean
  currentSong: Song | null
  coverUrl: string | null
  isPlaying: boolean
  lyrics: LyricLine[]
  hasLyrics: boolean
  currentTime: number
}>()

const emit = defineEmits<{
  close: []
  seek: [time: number]
}>()

const {
  bgOpacity,
  isTopChromeVisible,
  isMaximised,
  showTopChrome,
  handleTopChromeLeave,
  runEnterTransition,
  runLeaveTransition,
  updateMaximizeState,
  staggerStyle,
  minimize,
  toggleMaximize,
  closeApp,
} = usePlayerDetail()

const detailLeftRef = ref<InstanceType<typeof PlayerDetailLeft> | null>(null)
const isVisible = ref(false)
const isExpanded = ref(false)
const showLyrics = ref(true)
const positionLyrics = ref(true)
const isAnimating = ref(false)
let detailAnimationId = 0

const coverElement = computed(() =>
  detailLeftRef.value?.$el?.querySelector('.cover-container') as HTMLElement | null
    || detailLeftRef.value?.$el,
)

const handleClose = () => emit('close')

onMounted(() => {
  updateMaximizeState()
})

function toggleLyrics() {
  if (!props.hasLyrics) return
  showLyrics.value = !showLyrics.value
  positionLyrics.value = showLyrics.value
}

watch(() => props.hasLyrics, (hasLyrics) => {
  if (props.show) {
    showLyrics.value = hasLyrics
    positionLyrics.value = hasLyrics
  }
})

watch(() => props.show, async (visible) => {
  const animId = ++detailAnimationId

  if (visible) {
    isAnimating.value = true
    showLyrics.value = props.hasLyrics
    positionLyrics.value = props.hasLyrics
    isTopChromeVisible.value = true
    showTopChrome()
    isVisible.value = true
    isExpanded.value = true

    await nextTick()
    await new Promise(resolve => requestAnimationFrame(resolve))
    await runEnterTransition(coverElement.value)
    if (animId !== detailAnimationId) return
    isAnimating.value = false
  } else {
    isAnimating.value = true
    showLyrics.value = false
    isTopChromeVisible.value = false
    await runLeaveTransition(coverElement.value)
    if (animId !== detailAnimationId) return
    isVisible.value = false
    isExpanded.value = false
    isAnimating.value = false
    positionLyrics.value = false
  }
})
</script>

<template>
  <div
    class="player-detail"
    :class="{ visible: isVisible }"
  >
    <div class="player-inner">
      <div
        class="bg-wrapper"
        :style="{ opacity: bgOpacity }"
      >
        <PlayerDetailBackground :cover-url="props.coverUrl" :active="isVisible" />
        <div class="bg-fallback"></div>
      </div>

      <PlayerDetailTopChrome
        :is-visible="isVisible"
        :is-top-chrome-visible="isTopChromeVisible"
        :is-maximised="isMaximised"
        :current-song="props.currentSong"
        :stagger-style="(phase, dir, dist) => staggerStyle(props.show, phase, dir, dist)"
        @close="handleClose"
        @minimize="minimize"
        @toggle-maximize="toggleMaximize"
        @close-app="closeApp"
        @show-top-chrome="showTopChrome"
        @top-chrome-leave="handleTopChromeLeave"
      />

      <PlayerDetailLeft
        ref="detailLeftRef"
        :cover-url="props.coverUrl"
        :is-playing="props.isPlaying"
        :is-expanded="isExpanded"
        :show-lyrics="positionLyrics"
        :is-animating="isAnimating"
        @toggle-lyrics="toggleLyrics"
      />

      <PlayerDetailLyrics
        :lyrics="props.lyrics"
        :current-time="props.currentTime"
        :show="showLyrics"
        :is-playing="props.isPlaying"
        @seek="emit('seek', $event)"
      />
    </div>
  </div>
</template>

<style scoped>
.player-detail {
  position: fixed;
  inset: 0;
  z-index: 50;
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
  font-family: sans-serif;
  user-select: none;
  color: white;
  opacity: 0;
  pointer-events: none;
  visibility: hidden;
  transition: none;
}

.player-detail.visible {
  opacity: 1;
  pointer-events: auto;
  visibility: visible;
  transition: none;
}

.player-inner {
  position: relative;
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100%;
}

.bg-wrapper {
  position: absolute;
  inset: 0;
  transition: opacity 350ms cubic-bezier(0.4, 0, 0.2, 1);
}

.bg-fallback {
  position: absolute;
  inset: 0;
  z-index: -1;
  background: #0a0a0a;
}
</style>
