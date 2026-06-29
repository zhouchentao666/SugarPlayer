import { ref, computed, type CSSProperties } from 'vue'

interface Rect {
  x: number
  y: number
  width: number
  height: number
}

type AnimationPhase = 'idle' | 'entering' | 'leaving'

const DURATION = 500
const EASING = 'cubic-bezier(0.4, 0.0, 0.2, 1)'

export const animationPhase = ref<AnimationPhase>('idle')
export const isAnimating = ref(false)
export const footerCoverVisible = ref(true)
export const bgOpacity = ref(0)
export const staggerPhase = ref(0)
const firstRect = ref<Rect | null>(null)
const flipTransform = ref('')
const flipBorderRadius = ref('')
const flipTransition = ref('')

let currentAnimationId = 0
let staggerTimers: ReturnType<typeof setTimeout>[] = []

function clearStaggerTimers() {
  staggerTimers.forEach(t => clearTimeout(t))
  staggerTimers = []
}

export function captureFirst(el: HTMLElement) {
  const rect = el.getBoundingClientRect()
  firstRect.value = {
    x: rect.left,
    y: rect.top,
    width: rect.width,
    height: rect.height,
  }
}

function resetState() {
  animationPhase.value = 'idle'
  isAnimating.value = false
  footerCoverVisible.value = false
  staggerPhase.value = 0
  flipTransform.value = ''
  flipBorderRadius.value = ''
  flipTransition.value = ''
  // Note: bgOpacity is intentionally not reset here.
  // playEnter/playLeave already set it to the correct final value (1 or 0).
}

export function playEnter(lastEl: HTMLElement): Promise<void> {
  return new Promise((resolve) => {
    const animId = ++currentAnimationId
    animationPhase.value = 'entering'
    isAnimating.value = true
    footerCoverVisible.value = false
    bgOpacity.value = 0
    staggerPhase.value = 0
    clearStaggerTimers()

    if (!firstRect.value) {
      bgOpacity.value = 1
      staggerPhase.value = 3
      setTimeout(() => {
        resetState()
        resolve()
      }, DURATION)
      return
    }

    const lRect = lastEl.getBoundingClientRect()
    const dx = firstRect.value.x - lRect.left
    const dy = firstRect.value.y - lRect.top
    const sx = firstRect.value.width / lRect.width
    const sy = firstRect.value.height / lRect.height

    flipTransition.value = 'none'
    flipTransform.value = `translate(${dx}px, ${dy}px) scale(${sx}, ${sy})`
    flipBorderRadius.value = '8px'

    requestAnimationFrame(() => {
      if (animId !== currentAnimationId) { resolve(); return }
      requestAnimationFrame(() => {
        if (animId !== currentAnimationId) { resolve(); return }

        flipTransition.value = `transform ${DURATION}ms ${EASING}, border-radius ${DURATION}ms ${EASING}`
        flipTransform.value = 'translate(0, 0) scale(1, 1)'
        flipBorderRadius.value = '16px'

        bgOpacity.value = 1

        staggerTimers.push(
          setTimeout(() => {
            if (animId !== currentAnimationId) return
            staggerPhase.value = 1
          }, DURATION * 0.45),

          setTimeout(() => {
            if (animId !== currentAnimationId) return
            staggerPhase.value = 2
          }, DURATION * 0.6),

          setTimeout(() => {
            if (animId !== currentAnimationId) return
            staggerPhase.value = 3
          }, DURATION * 0.8),
        )

        setTimeout(() => {
          if (animId !== currentAnimationId) { resolve(); return }
          resetState()
          resolve()
        }, DURATION)
      })
    })
  })
}

export function playLeave(detailCoverEl: HTMLElement): Promise<void> {
  return new Promise((resolve) => {
    const animId = ++currentAnimationId
    animationPhase.value = 'leaving'
    isAnimating.value = true
    clearStaggerTimers()
    staggerPhase.value = 0
    footerCoverVisible.value = true

    if (!firstRect.value) {
      bgOpacity.value = 0
      setTimeout(() => {
        resetState()
        resolve()
      }, DURATION * 0.6)
      return
    }

    const footerEl = document.querySelector('[data-footer-cover]') as HTMLElement | null
    if (footerEl) {
      const rect = footerEl.getBoundingClientRect()
      firstRect.value = {
        x: rect.left,
        y: rect.top,
        width: rect.width,
        height: rect.height,
      }
    }

    const lRect = detailCoverEl.getBoundingClientRect()
    const dx = firstRect.value.x - lRect.left
    const dy = firstRect.value.y - lRect.top
    const sx = firstRect.value.width / lRect.width
    const sy = firstRect.value.height / lRect.height

    flipTransition.value = 'none'
    flipTransform.value = 'translate(0, 0) scale(1, 1)'
    flipBorderRadius.value = '16px'
    bgOpacity.value = 0

    requestAnimationFrame(() => {
      if (animId !== currentAnimationId) { resolve(); return }
      requestAnimationFrame(() => {
        if (animId !== currentAnimationId) { resolve(); return }

        flipTransition.value = `transform ${DURATION}ms ${EASING}, border-radius ${DURATION}ms ${EASING}`
        flipTransform.value = `translate(${dx}px, ${dy}px) scale(${sx}, ${sy})`
        flipBorderRadius.value = '8px'

        setTimeout(() => {
          if (animId !== currentAnimationId) { resolve(); return }
          resetState()
          resolve()
        }, DURATION)
      })
    })
  })
}

export function cancel() {
  currentAnimationId++
  clearStaggerTimers()
  resetState()
}

export const coverStyle = computed<CSSProperties>(() => {
  const style: CSSProperties = {}
  if (flipTransform.value) {
    style.transform = flipTransform.value
    style.transformOrigin = 'top left'
  }
  if (flipBorderRadius.value) {
    style.borderRadius = flipBorderRadius.value
  }
  if (flipTransition.value && flipTransition.value !== 'none') {
    style.transition = flipTransition.value
  } else if (flipTransition.value === 'none') {
    style.transition = 'none'
  }
  return style
})

export function useSharedTransition() {
  return {
    animationPhase,
    isAnimating,
    coverStyle,
    footerCoverVisible,
    bgOpacity,
    staggerPhase,
    captureFirst,
    playEnter,
    playLeave,
    cancel,
    DURATION,
  }
}