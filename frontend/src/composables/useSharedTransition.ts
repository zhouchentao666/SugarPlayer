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

function parseTransformMatrix(el: HTMLElement) {
  const style = window.getComputedStyle(el).transform
  if (!style || style === 'none') {
    return { x: 0, y: 0, sx: 1, sy: 1 }
  }
  const matrix = new DOMMatrix(style)
  const sx = Math.sqrt(matrix.a ** 2 + matrix.b ** 2)
  const sy = Math.sqrt(matrix.c ** 2 + matrix.d ** 2)
  return { x: matrix.e, y: matrix.f, sx, sy }
}

function naturalRectOf(el: HTMLElement): Rect {
  const matrix = parseTransformMatrix(el)
  const rect = el.getBoundingClientRect()
  return {
    x: rect.left - matrix.x,
    y: rect.top - matrix.y,
    width: rect.width / matrix.sx,
    height: rect.height / matrix.sy,
  }
}

function buildTransform(from: Rect, to: Rect): string {
  const dx = to.x - from.x
  const dy = to.y - from.y
  const sx = to.width / from.width
  const sy = to.height / from.height
  return `translate(${dx}px, ${dy}px) scale(${sx}, ${sy})`
}

function resetState() {
  animationPhase.value = 'idle'
  isAnimating.value = false
  staggerPhase.value = 0
  flipTransform.value = ''
  flipBorderRadius.value = ''
  flipTransition.value = ''
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

export function playEnter(lastEl: HTMLElement): Promise<void> {
  return new Promise((resolve) => {
    const animId = ++currentAnimationId
    const wasAnimating = isAnimating.value
    animationPhase.value = 'entering'
    isAnimating.value = true
    footerCoverVisible.value = false
    staggerPhase.value = 0
    clearStaggerTimers()

    if (!wasAnimating) {
      bgOpacity.value = 0
    }

    const naturalRect = naturalRectOf(lastEl)

    flipTransition.value = 'none'
    if (wasAnimating) {
      const computedTransform = window.getComputedStyle(lastEl).transform
      const computedBorderRadius = window.getComputedStyle(lastEl).borderRadius
      flipTransform.value = computedTransform === 'none' ? 'translate(0, 0) scale(1, 1)' : computedTransform
      flipBorderRadius.value = computedBorderRadius
    } else if (firstRect.value) {
      flipTransform.value = buildTransform(naturalRect, firstRect.value)
      flipBorderRadius.value = '8px'
    } else {
      bgOpacity.value = 1
      staggerPhase.value = 3
      setTimeout(() => {
        resetState()
        resolve()
      }, DURATION)
      return
    }

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
    const wasAnimating = isAnimating.value
    animationPhase.value = 'leaving'
    isAnimating.value = true
    clearStaggerTimers()
    staggerPhase.value = 0
    footerCoverVisible.value = false

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

    if (!firstRect.value) {
      bgOpacity.value = 0
      setTimeout(() => {
        resetState()
        footerCoverVisible.value = true
        resolve()
      }, DURATION * 0.6)
      return
    }

    const naturalRect = naturalRectOf(detailCoverEl)

    flipTransition.value = 'none'
    if (wasAnimating) {
      const computedTransform = window.getComputedStyle(detailCoverEl).transform
      const computedBorderRadius = window.getComputedStyle(detailCoverEl).borderRadius
      flipTransform.value = computedTransform === 'none' ? 'translate(0, 0) scale(1, 1)' : computedTransform
      flipBorderRadius.value = computedBorderRadius
    } else {
      flipTransform.value = 'translate(0, 0) scale(1, 1)'
      flipBorderRadius.value = '16px'
    }
    bgOpacity.value = 0

    requestAnimationFrame(() => {
      if (animId !== currentAnimationId) { resolve(); return }
      requestAnimationFrame(() => {
        if (animId !== currentAnimationId) { resolve(); return }

        flipTransition.value = `transform ${DURATION}ms ${EASING}, border-radius ${DURATION}ms ${EASING}`
        flipTransform.value = buildTransform(naturalRect, firstRect.value!)
        flipBorderRadius.value = '8px'

        setTimeout(() => {
          if (animId !== currentAnimationId) { resolve(); return }
          resetState()
          footerCoverVisible.value = true
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
  footerCoverVisible.value = true
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
