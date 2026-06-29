import type { Song } from '../types'

export function formatDuration(seconds: number): string {
  if (!seconds || seconds < 0) return '--:--'
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

export function displayTitle(song: Song): string {
  return song.metadata?.title || song.title
}

export function displayArtist(song: Song): string {
  return song.metadata?.artist || '未知艺术家'
}

export function displayAlbum(song: Song): string {
  return song.metadata?.album || ''
}

export function displayDuration(song: Song): string {
  return formatDuration(song.metadata?.duration ?? 0)
}
