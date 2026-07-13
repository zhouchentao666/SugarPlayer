import { ref } from 'vue'
import type { OnlineSong } from '../../bindings/sugarplayer/models'

// Holds the currently playing online song so that lyric loading (useLyrics)
// can resolve lyrics for it independently of the local-file path logic.
export const currentOnlineSong = ref<OnlineSong | null>(null)
