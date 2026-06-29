import { computed, onBeforeUnmount, ref } from 'vue'
import { Window, Application } from '@wailsio/runtime'
import { useSharedTransition, bgOpacity, staggerPhase } from './useSharedTransition'

export function usePlayerDetail() {
  const { captureFirst, playEnter, playLeave, cancel } = useSharedTransition()
  const isMaximised = ref(false)

  async function updateMaximizeState() {
    isMaximised.value = await Window.IsMaximised()
  }

  async function toggleMaximize() {
    Window.ToggleMaximise()
    await updateMaximizeState()
  }

  const TOP_CHROME_HIDE_DELAY = 2500
  const isTopChromeVisible = ref(false)
  let topChromeHideTimer: ReturnType<typeof setTimeout> | null = null

  const clearTopChromeHideTimer = () => {
    if (topChromeHideTimer) {
      clearTimeout(topChromeHideTimer)
      topChromeHideTimer = null
    }
  }

  const scheduleTopChromeHide = () => {
    clearTopChromeHideTimer()
    topChromeHideTimer = setTimeout(() => {
      isTopChromeVisible.value = false
      topChromeHideTimer = null
    }, TOP_CHROME_HIDE_DELAY)
  }

  const showTopChrome = () => {
    clearTopChromeHideTimer()
    isTopChromeVisible.value = true
  }

  const handleTopChromeLeave = () => {
    scheduleTopChromeHide()
  }

  const runEnterTransition = async (detailCover: HTMLElement | null | undefined) => {
    const footerCover = document.querySelector('[data-footer-cover]') as HTMLElement | null
    if (footerCover) captureFirst(footerCover)

    if (detailCover) {
      await playEnter(detailCover)
    }
  }

  const runLeaveTransition = async (detailCover: HTMLElement | null | undefined) => {
    if (detailCover) {
      await playLeave(detailCover)
    }
  }

  const staggerStyle = (show: boolean, phase: number, translateDir: 'Y' | 'X' = 'Y', distance = 20) => {
    const visible = show || staggerPhase.value >= phase
    const translate = translateDir === 'Y' ? `translateY(${distance}px)` : `translateX(${distance}px)`

    return {
      opacity: visible ? 1 : 0,
      transform: visible ? 'translate(0, 0)' : translate,
      transition: `opacity 400ms cubic-bezier(0.22,1,0.36,1) ${show ? phase * 100 : 0}ms, transform 400ms cubic-bezier(0.22,1,0.36,1) ${show ? phase * 100 : 0}ms`,
    }
  }

  const minimize = () => Window.Minimise()
  const closeApp = () => Application.Quit()

  onBeforeUnmount(() => {
    clearTopChromeHideTimer()
  })

  return {
    bgOpacity: computed(() => bgOpacity.value),
    staggerPhase: computed(() => staggerPhase.value),
    isTopChromeVisible,
    isMaximised,
    showTopChrome,
    handleTopChromeLeave,
    runEnterTransition,
    runLeaveTransition,
    cancel,
    updateMaximizeState,
    staggerStyle,
    minimize,
    toggleMaximize,
    closeApp,
  }
}
