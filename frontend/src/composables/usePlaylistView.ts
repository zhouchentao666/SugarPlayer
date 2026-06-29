import { ref, computed, watch, type Ref } from 'vue'
import type { Playlist, Song } from '../types'

export type SortMode = 'default' | 'title' | 'artist' | 'album' | 'duration'
export type ViewMode = 'list' | 'grid'

const sortLabels: Record<SortMode, string> = {
  default: '默认排序',
  title: '按标题',
  artist: '按艺术家',
  album: '按专辑',
  duration: '按时长',
}

export function usePlaylistView(playlist: Ref<Playlist>) {
  const searchQuery = ref('')
  const sortMode = ref<SortMode>('default')
  const viewMode = ref<ViewMode>('list')
  const batchMode = ref(false)
  const selectedIds = ref<Set<string>>(new Set())

  const displaySongs = computed(() => {
    let songs = [...playlist.value.songs]
    const query = searchQuery.value.trim().toLowerCase()
    if (query) {
      songs = songs.filter((song) => {
        const text = [
          song.metadata?.title,
          song.metadata?.artist,
          song.metadata?.album,
          song.title,
        ]
          .filter(Boolean)
          .join(' ')
          .toLowerCase()
        return text.includes(query)
      })
    }
    switch (sortMode.value) {
      case 'title':
        songs.sort((a, b) => (a.metadata?.title || a.title).localeCompare(b.metadata?.title || b.title))
        break
      case 'artist':
        songs.sort((a, b) => (a.metadata?.artist || '').localeCompare(b.metadata?.artist || ''))
        break
      case 'album':
        songs.sort((a, b) => (a.metadata?.album || '').localeCompare(b.metadata?.album || ''))
        break
      case 'duration':
        songs.sort((a, b) => (a.metadata?.duration ?? 0) - (b.metadata?.duration ?? 0))
        break
    }
    return songs
  })

  const selectedSongs = computed(() =>
    displaySongs.value.filter(song => selectedIds.value.has(song.id))
  )

  const allSelected = computed(() =>
    displaySongs.value.length > 0 && displaySongs.value.every(song => selectedIds.value.has(song.id))
  )

  function toggleSelection(id: string) {
    const next = new Set(selectedIds.value)
    if (next.has(id)) next.delete(id)
    else next.add(id)
    selectedIds.value = next
  }

  function selectAll() {
    selectedIds.value = new Set(displaySongs.value.map(song => song.id))
  }

  function clearSelection() {
    selectedIds.value = new Set()
  }

  function exitBatchMode() {
    batchMode.value = false
    clearSelection()
  }

  watch(() => playlist.value.id, exitBatchMode)

  return {
    searchQuery,
    sortMode,
    viewMode,
    batchMode,
    selectedIds,
    displaySongs,
    selectedSongs,
    allSelected,
    sortLabels,
    toggleSelection,
    selectAll,
    clearSelection,
    exitBatchMode,
  }
}
