<script lang="ts" setup>
import { ref, onMounted, computed, watch } from 'vue'
import {
  OnlineSearch,
  OnlineSources,
  SwitchSongSource,
  OnlineRecommendPlaylists,
  OnlineUserPlaylists,
  OnlineSearchCollections,
  OnlineCollectionSongs,
} from '../../bindings/sugarplayer/app'
import type { OnlineSong, OnlineSource, OnlineCollection } from '../../bindings/sugarplayer/models'
import type { Song } from '../types'
import OnlineDownloadDialog from './OnlineDownloadDialog.vue'

const props = defineProps<{
  currentSong: Song | null
  autoSwitch: boolean
  pinnedCollections: OnlineCollection[]
  openCollection: OnlineCollection | null
}>()

const emit = defineEmits<{
  (e: 'play', list: OnlineSong[], index: number): void
  (e: 'favorite', song: OnlineSong): void
  (e: 'toggle-pin', collection: OnlineCollection): void
}>()

const favorited = ref<Set<string>>(new Set())
const downloadSong = ref<OnlineSong | null>(null)
const showDownload = ref(false)
const toast = ref('')
let toastTimer: ReturnType<typeof setTimeout> | null = null

function showToast(msg: string) {
  toast.value = msg
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => (toast.value = ''), 1800)
}

function favoriteKey(song: OnlineSong): string {
  return `${song.source}:${song.id}`
}

function isFavorited(song: OnlineSong): boolean {
  return favorited.value.has(favoriteKey(song))
}

function toggleFavorite(song: OnlineSong) {
  const key = favoriteKey(song)
  if (favorited.value.has(key)) return
  favorited.value = new Set(favorited.value).add(key)
  emit('favorite', song)
  showToast('已添加到「我的喜欢」')
}

function openDownload(song: OnlineSong) {
  downloadSong.value = song
  showDownload.value = true
}

// ---------- 搜索 / 视图状态 ----------
const keyword = ref('')
const searchType = ref<'song' | 'playlist' | 'album'>('song')
const allSources = ref<OnlineSource[]>([])
const selected = ref<string[]>([])

const mode = ref<'discover' | 'list'>('discover')
const listKind = ref<'songs' | 'collections'>('songs')
const results = ref<OnlineSong[]>([])
const collections = ref<OnlineCollection[]>([])
const currentCollection = ref<OnlineCollection | null>(null)
const listTitle = ref('')
const discoverOrigin = ref(false)

const loading = ref(false)
const loaded = ref(false)
const error = ref('')

const sourceNameMap = computed(() => {
  const m: Record<string, string> = {}
  for (const s of allSources.value) m[s.id] = s.name
  return m
})

const searchPlaceholder = computed(() => {
  if (searchType.value === 'playlist') return '搜索歌单，或粘贴歌单分享链接'
  if (searchType.value === 'album') return '搜索专辑，或粘贴专辑分享链接'
  return '搜索歌曲、歌手，或粘贴分享链接'
})

onMounted(async () => {
  try {
    const sources = await OnlineSources()
    allSources.value = sources
    selected.value = sources.filter(s => s.enabled).map(s => s.id)
  } catch {
    allSources.value = []
  }
})

function toggleSource(id: string) {
  const i = selected.value.indexOf(id)
  if (i >= 0) selected.value.splice(i, 1)
  else selected.value.push(id)
}

function selectAll() {
  selected.value = allSources.value.map(s => s.id)
}

function clearAll() {
  selected.value = []
}

const resolving = ref(false)
const resolveProgress = ref({ done: 0, total: 0 })

// 为搜索结果中的无效音源寻找可用替代音源（批量换源）。
async function batchResolve() {
  const list = results.value
  if (resolving.value || list.length === 0) return
  resolving.value = true
  resolveProgress.value = { done: 0, total: list.length }
  const concurrency = 6
  let cursor = 0
  const worker = async () => {
    while (cursor < list.length) {
      const i = cursor++
      try {
        const alt = await SwitchSongSource(list[i])
        if (alt.source !== list[i].source || alt.id !== list[i].id) {
          results.value[i] = alt
        }
      } catch {
        // 换源失败，保持原结果
      } finally {
        resolveProgress.value.done++
      }
    }
  }
  const workers = Array.from({ length: Math.min(concurrency, list.length) }, () => worker())
  await Promise.all(workers)
  resolving.value = false
}

async function search() {
  const q = keyword.value.trim()
  if (!q) return
  loading.value = true
  error.value = ''
  loaded.value = false
  mode.value = 'list'
  discoverOrigin.value = false
  try {
    if (searchType.value === 'song') {
      listKind.value = 'songs'
      currentCollection.value = null
      listTitle.value = `“${q}” 的搜索结果`
      const list = await OnlineSearch(q, selected.value)
      results.value = list
      if (props.autoSwitch) batchResolve()
    } else {
      listKind.value = 'collections'
      currentCollection.value = null
      listTitle.value = `“${q}” 的${searchType.value === 'album' ? '专辑' : '歌单'}`
      collections.value = await OnlineSearchCollections(q, searchType.value, selected.value)
    }
    loaded.value = true
  } catch (e) {
    results.value = []
    collections.value = []
    error.value = e instanceof Error ? e.message : '搜索失败'
  } finally {
    loading.value = false
  }
}

function onSearchInputEnter(e: KeyboardEvent) {
  if (e.key === 'Enter') search()
}

// ---------- 发现页：每日推荐 / 我的歌单 ----------
async function loadRecommend() {
  loading.value = true
  error.value = ''
  loaded.value = false
  mode.value = 'list'
  listKind.value = 'collections'
  currentCollection.value = null
  discoverOrigin.value = true
  listTitle.value = '每日推荐歌单'
  try {
    collections.value = await OnlineRecommendPlaylists([])
    loaded.value = true
  } catch (e) {
    collections.value = []
    error.value = e instanceof Error ? e.message : '加载失败'
  } finally {
    loading.value = false
  }
}

async function loadUserPlaylists() {
  loading.value = true
  error.value = ''
  loaded.value = false
  mode.value = 'list'
  listKind.value = 'collections'
  currentCollection.value = null
  discoverOrigin.value = true
  listTitle.value = '我的歌单'
  try {
    collections.value = await OnlineUserPlaylists([])
    loaded.value = true
  } catch (e) {
    collections.value = []
    error.value = e instanceof Error ? e.message : '加载失败'
  } finally {
    loading.value = false
  }
}

// ---------- 打开歌单 / 专辑详情 ----------
async function openCollectionSongs(col: OnlineCollection) {
  loading.value = true
  error.value = ''
  loaded.value = false
  mode.value = 'list'
  listKind.value = 'songs'
  currentCollection.value = col
  listTitle.value = col.name
  try {
    results.value = await OnlineCollectionSongs(col)
    loaded.value = true
  } catch (e) {
    results.value = []
    error.value = e instanceof Error ? e.message : '加载失败'
  } finally {
    loading.value = false
  }
}

function goBack() {
  if (currentCollection.value) {
    currentCollection.value = null
    listKind.value = 'collections'
  } else {
    mode.value = 'discover'
    discoverOrigin.value = false
  }
}

// 从侧栏点击固定歌单时由父组件通过 openCollection 传入
watch(
  () => props.openCollection,
  (col) => {
    if (col) {
      discoverOrigin.value = false
      openCollectionSongs(col)
    }
  },
  { immediate: true }
)

// ---------- 固定到侧栏 ----------
function isPinned(col: OnlineCollection): boolean {
  return props.pinnedCollections.some(
    p => p.source === col.source && p.id === col.id && p.kind === col.kind
  )
}

function togglePin(col: OnlineCollection) {
  emit('toggle-pin', col)
  showToast(isPinned(col) ? '已取消固定' : '已固定到侧栏')
}

// ---------- 播放 ----------
function playAll() {
  if (results.value.length) emit('play', results.value, 0)
}

function playIndex(index: number) {
  emit('play', results.value, index)
}

function isPlaying(song: OnlineSong): boolean {
  return props.currentSong?.id === `online:${song.source}:${song.id}`
}

function formatDuration(seconds: number): string {
  if (!seconds || seconds < 0) return '--:--'
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}
</script>

<template>
  <div class="online-view">
    <div class="search-panel">
      <div class="search-row">
        <div class="seg">
          <button :class="['seg-item', { active: searchType === 'song' }]" @click="searchType = 'song'">单曲</button>
          <button :class="['seg-item', { active: searchType === 'playlist' }]" @click="searchType = 'playlist'">歌单</button>
          <button :class="['seg-item', { active: searchType === 'album' }]" @click="searchType = 'album'">专辑</button>
        </div>
        <div class="search-box">
          <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="11" cy="11" r="7" />
            <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
          <input
            v-model="keyword"
            class="search-input"
            type="text"
            :placeholder="searchPlaceholder"
            @keydown="onSearchInputEnter"
          />
          <button class="search-btn" :disabled="loading || !keyword.trim()" @click="search">
            {{ loading ? '搜索中…' : '搜索' }}
          </button>
        </div>
      </div>

      <div class="source-bar">
        <div class="source-chips">
          <button
            v-for="src in allSources"
            :key="src.id"
            :class="['source-chip', { active: selected.includes(src.id) }]"
            @click="toggleSource(src.id)"
          >
            {{ src.name }}
          </button>
        </div>
        <div class="source-actions">
          <button class="link-btn" @click="selectAll">全选</button>
          <span class="divider">·</span>
          <button class="link-btn" @click="clearAll">清空</button>
        </div>
      </div>
    </div>

    <div class="results">
      <!-- 发现页 -->
      <div v-if="mode === 'discover'" class="discover">
        <h2 class="discover-title">发现</h2>
        <div class="discover-grid">
          <button class="discover-card" @click="loadRecommend">
            <span class="discover-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
                <path d="M12 2v4M12 18v4M4.9 4.9l2.8 2.8M16.3 16.3l2.8 2.8M2 12h4M18 12h4M4.9 19.1l2.8-2.8M16.3 7.7l2.8-2.8" />
                <circle cx="12" cy="12" r="3.2" />
              </svg>
            </span>
            <span class="discover-name">每日推荐歌单</span>
            <span class="discover-desc">聚合网易云 / QQ / 酷狗 / 酷我 的每日推荐</span>
          </button>
          <button class="discover-card" @click="loadUserPlaylists">
            <span class="discover-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
                <path d="M9 18V5l12-2v13" />
                <circle cx="6" cy="18" r="3" />
                <circle cx="18" cy="16" r="3" />
              </svg>
            </span>
            <span class="discover-name">我的歌单</span>
            <span class="discover-desc">登录后查看个人创建 / 收藏的歌单</span>
          </button>
        </div>
        <p class="discover-hint">提示：在上方切换「歌单 / 专辑」可直接搜索；把喜欢的歌单固定到侧栏可随时打开。</p>
      </div>

      <!-- 加载 / 错误 -->
      <div v-else-if="loading" class="state">
        <div class="spinner"></div>
        <span>正在加载…</span>
      </div>
      <div v-else-if="error" class="state error">
        <span>{{ error }}</span>
      </div>

      <!-- 歌单 / 专辑 网格 -->
      <template v-else-if="listKind === 'collections'">
        <div class="list-toolbar">
          <button class="back-btn" @click="goBack">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="15 18 9 12 15 6" />
            </svg>
            返回
          </button>
          <span class="list-title">{{ listTitle }}</span>
          <span class="result-count">共 {{ collections.length }} 个</span>
        </div>
        <div v-if="loaded && collections.length === 0" class="state">
          <span>没有找到相关{{ searchType === 'album' ? '专辑' : '歌单' }}，换个关键词或音源试试</span>
        </div>
        <div v-else class="collection-grid">
          <div
            v-for="col in collections"
            :key="col.source + ':' + col.kind + ':' + col.id"
            class="collection-card"
            @click="openCollectionSongs(col)"
          >
            <div class="collection-cover">
              <img v-if="col.cover" :src="col.cover" class="cover-img" alt="" loading="lazy" />
              <div v-else class="cover-fallback">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                  <path d="M9 18V5l12-2v13" />
                  <circle cx="6" cy="18" r="3" />
                  <circle cx="18" cy="16" r="3" />
                </svg>
              </div>
              <button
                class="pin-btn"
                :class="{ active: isPinned(col) }"
                :title="isPinned(col) ? '取消固定' : '固定到侧栏'"
                @click.stop="togglePin(col)"
              >
                <svg viewBox="0 0 24 24" :fill="isPinned(col) ? 'currentColor' : 'none'" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M12 17v5M9 10.8V4h6v6.8l2 2.2H7l2-2.2z" />
                </svg>
              </button>
            </div>
            <div class="collection-meta">
              <div class="collection-name">{{ col.name }}</div>
              <div class="collection-sub">
                <span class="source-badge">{{ sourceNameMap[col.source] || col.source }}</span>
                <span v-if="col.creator" class="collection-creator">{{ col.creator }}</span>
                <span v-if="col.trackCount" class="collection-count">{{ col.trackCount }} 首</span>
              </div>
            </div>
          </div>
        </div>
      </template>

      <!-- 歌曲列表 -->
      <template v-else>
        <div class="list-toolbar">
          <button v-if="currentCollection" class="back-btn" @click="goBack">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="15 18 9 12 15 6" />
            </svg>
            返回
          </button>
          <span class="list-title">{{ listTitle }}</span>
          <button v-if="currentCollection" class="pin-btn inline" :class="{ active: isPinned(currentCollection) }" :title="isPinned(currentCollection) ? '取消固定' : '固定到侧栏'" @click="togglePin(currentCollection)">
            <svg viewBox="0 0 24 24" :fill="isPinned(currentCollection) ? 'currentColor' : 'none'" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M12 17v5M9 10.8V4h6v6.8l2 2.2H7l2-2.2z" />
            </svg>
            {{ isPinned(currentCollection) ? '已固定' : '固定到侧栏' }}
          </button>
          <span v-if="!currentCollection" class="result-count">共 {{ results.length }} 条结果</span>
          <span v-else class="result-count">共 {{ results.length }} 首</span>
        </div>

        <div v-if="loaded && results.length === 0" class="state">
          <span>这个{{ currentCollection?.kind === 'album' ? '专辑' : '歌单' }}暂时没有可播放的歌曲</span>
        </div>

        <template v-else>
          <div v-if="!currentCollection" class="result-toolbar">
            <button class="play-all" @click="playAll">
              <svg viewBox="0 0 24 24" fill="currentColor">
                <path d="M8 5v14l11-7z" />
              </svg>
              播放全部
            </button>
            <button class="switch-btn" :disabled="resolving || results.length === 0" @click="batchResolve">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="23 4 23 10 17 10" />
                <path d="M20.5 15a9 9 0 1 1-2.1-9.4L23 10" />
              </svg>
              {{ resolving ? `换源中 ${resolveProgress.done}/${resolveProgress.total}` : '批量换源' }}
            </button>
          </div>

          <div class="song-list">
            <div class="list-header">
              <span class="col-cover"></span>
              <span class="col-index">#</span>
              <span class="col-title">标题</span>
              <span class="col-artist">艺术家</span>
              <span class="col-album">专辑</span>
              <span class="col-duration">时长</span>
              <span class="col-action"></span>
            </div>
            <div
              v-for="(song, index) in results"
              :key="song.source + ':' + song.id"
              :class="['song-item', { playing: isPlaying(song) }]"
              @dblclick="playIndex(index)"
            >
              <div class="col-cover">
                <img
                  v-if="song.cover"
                  :src="song.cover"
                  class="cover-img"
                  alt=""
                  loading="lazy"
                />
                <div v-else class="cover-fallback">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                    <path d="M9 18V5l12-2v13" />
                    <circle cx="6" cy="18" r="3" />
                    <circle cx="18" cy="16" r="3" />
                  </svg>
                </div>
                <button class="cover-play" title="播放" @click.stop="playIndex(index)">
                  <svg v-if="!isPlaying(song)" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M8 5v14l11-7z" />
                  </svg>
                  <svg v-else viewBox="0 0 24 24" fill="currentColor">
                    <path d="M6 5h4v14H6zM14 5h4v14h-4z" />
                  </svg>
                </button>
              </div>
              <span class="col-index">{{ index + 1 }}</span>
              <div class="col-title">
                <div class="primary-text">{{ song.name }}</div>
                <div class="secondary-text">{{ song.artist }}</div>
              </div>
              <span class="col-artist secondary-text">{{ song.artist }}</span>
              <span class="col-album secondary-text">{{ song.album || '—' }}</span>
              <span class="col-duration secondary-text">{{ formatDuration(song.duration) }}</span>
              <div class="col-action">
                <button
                  class="icon-btn"
                  :class="{ active: isFavorited(song) }"
                  title="收藏到我的喜欢"
                  @click.stop="toggleFavorite(song)"
                >
                  <svg viewBox="0 0 24 24" :fill="isFavorited(song) ? 'currentColor' : 'none'" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M20.8 4.6a5.5 5.5 0 0 0-7.8 0L12 5.6l-1-1a5.5 5.5 0 0 0-7.8 7.8l1 1L12 21l7.8-7.6 1-1a5.5 5.5 0 0 0 0-7.8z" />
                  </svg>
                </button>
                <button class="icon-btn" title="下载" @click.stop="openDownload(song)">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                    <polyline points="7 10 12 15 17 10" />
                    <line x1="12" y1="15" x2="12" y2="3" />
                  </svg>
                </button>
                <span class="source-badge">{{ sourceNameMap[song.source] || song.source }}</span>
              </div>
            </div>
          </div>
        </template>
      </template>

      <OnlineDownloadDialog
        :show="showDownload"
        :song="downloadSong"
        @close="showDownload = false"
      />

      <Transition name="toast-fade">
        <div v-if="toast" class="toast">{{ toast }}</div>
      </Transition>
    </div>
  </div>
</template>

<style scoped>
.online-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.search-panel {
  padding: 20px 24px 12px;
  flex-shrink: 0;
}

.search-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
  padding: 0 14px;
  height: 44px;
  border-radius: 22px;
  background: var(--fluent-input-bg);
  border: 1px solid var(--fluent-input-border);
  transition: border-color 0.18s ease, box-shadow 0.18s ease;
}

.search-box:focus-within {
  border-color: var(--fluent-accent);
  box-shadow: 0 0 0 3px rgba(0, 120, 212, 0.2);
}

.search-icon {
  width: 18px;
  height: 18px;
  color: var(--fluent-text-secondary);
  flex-shrink: 0;
}

.search-input {
  flex: 1;
  height: 100%;
  border: none;
  background: transparent;
  color: var(--fluent-text);
  font-size: 14px;
  outline: none;
}

.search-input::placeholder {
  color: var(--fluent-text-secondary);
}

.search-btn {
  flex-shrink: 0;
  height: 32px;
  padding: 0 18px;
  border: none;
  border-radius: 16px;
  background: var(--fluent-accent);
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: filter 0.18s ease, opacity 0.18s ease;
}

.search-btn:hover:not(:disabled) {
  filter: brightness(1.1);
}

.search-btn:disabled {
  opacity: 0.5;
  cursor: default;
}

.source-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-top: 14px;
  flex-wrap: wrap;
}

.seg {
  display: flex;
  gap: 2px;
  padding: 3px;
  border-radius: 18px;
  background: var(--fluent-bg-active);
  flex-shrink: 0;
}

.seg-item {
  padding: 5px 14px;
  border: none;
  border-radius: 15px;
  background: transparent;
  color: var(--fluent-text-secondary);
  font-size: 12px;
  cursor: pointer;
  transition: background 0.18s ease, color 0.18s ease;
}

.seg-item.active {
  background: var(--fluent-bg-glass);
  color: var(--fluent-text);
  font-weight: 600;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.15);
}

.source-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  flex: 1;
}

.source-chip {
  padding: 5px 12px;
  border: 1px solid var(--fluent-border);
  border-radius: 14px;
  background: transparent;
  color: var(--fluent-text-secondary);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.18s ease;
}

.source-chip:hover {
  background: var(--fluent-bg-hover);
  color: var(--fluent-text);
}

.source-chip.active {
  background: var(--fluent-accent);
  border-color: var(--fluent-accent);
  color: #fff;
}

.source-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--fluent-text-secondary);
  font-size: 12px;
  flex-shrink: 0;
}

.link-btn {
  border: none;
  background: transparent;
  color: var(--fluent-accent);
  font-size: 12px;
  cursor: pointer;
  padding: 2px 4px;
}

.link-btn:hover {
  text-decoration: underline;
}

.divider {
  opacity: 0.5;
}

.results {
  flex: 1;
  overflow-y: auto;
  padding: 4px 24px 16px;
}

/* 发现页 */
.discover {
  padding-top: 8px;
}

.discover-title {
  margin: 4px 0 14px;
  font-size: 18px;
  font-weight: 700;
}

.discover-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 14px;
}

.discover-card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 20px;
  border: 1px solid var(--fluent-border);
  border-radius: 16px;
  background: var(--fluent-bg-glass);
  text-align: left;
  cursor: pointer;
  transition: background 0.18s ease, border-color 0.18s ease, transform 0.18s ease;
}

.discover-card:hover {
  background: var(--fluent-bg-hover);
  border-color: var(--fluent-accent);
  transform: translateY(-2px);
}

.discover-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: var(--fluent-accent);
  color: #fff;
}

.discover-icon svg {
  width: 24px;
  height: 24px;
}

.discover-name {
  font-size: 15px;
  font-weight: 600;
}

.discover-desc {
  font-size: 12px;
  color: var(--fluent-text-secondary);
  line-height: 1.5;
}

.discover-hint {
  margin: 18px 0 0;
  font-size: 12px;
  color: var(--fluent-text-secondary);
}

/* 列表工具条 */
.list-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 4px 12px;
}

.list-title {
  font-size: 15px;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  height: 32px;
  padding: 0 12px;
  border: 1px solid var(--fluent-border);
  border-radius: 16px;
  background: transparent;
  color: var(--fluent-text);
  font-size: 13px;
  cursor: pointer;
  transition: background 0.18s ease, border-color 0.18s ease;
  flex-shrink: 0;
}

.back-btn:hover {
  background: var(--fluent-bg-hover);
  border-color: var(--fluent-accent);
}

.back-btn svg {
  width: 16px;
  height: 16px;
}

/* 歌单 / 专辑 网格 */
.collection-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: 16px;
  padding-top: 4px;
}

.collection-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  border: none;
  background: transparent;
  cursor: pointer;
  text-align: left;
}

.collection-cover {
  position: relative;
  width: 100%;
  aspect-ratio: 1 / 1;
  border-radius: 12px;
  overflow: hidden;
  background: var(--fluent-bg-active);
}

.collection-cover .cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.cover-fallback {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--fluent-text-secondary);
}

.cover-fallback svg {
  width: 36px;
  height: 36px;
}

.pin-btn {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.5);
  color: #fff;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.18s ease, background 0.18s ease, color 0.18s ease;
}

.collection-cover:hover .pin-btn {
  opacity: 1;
}

.pin-btn:hover {
  background: rgba(0, 0, 0, 0.7);
}

.pin-btn.active {
  opacity: 1;
  background: var(--fluent-accent);
  color: #fff;
}

.pin-btn svg {
  width: 16px;
  height: 16px;
}

.pin-btn.inline {
  position: static;
  opacity: 1;
  width: auto;
  height: 32px;
  padding: 0 12px;
  gap: 6px;
  border-radius: 16px;
  background: var(--fluent-bg-active);
  color: var(--fluent-text);
  font-size: 12px;
}

.pin-btn.inline.active {
  background: var(--fluent-accent);
  color: #fff;
}

.pin-btn.inline svg {
  width: 15px;
  height: 15px;
}

.collection-meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.collection-name {
  font-size: 13px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.collection-sub {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
  font-size: 11px;
  color: var(--fluent-text-secondary);
}

.collection-creator {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.collection-count {
  white-space: nowrap;
}

/* 通用状态 */
.state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 60px 20px;
  color: var(--fluent-text-secondary);
  font-size: 13px;
  text-align: center;
}

.state.error {
  color: #ff8080;
}

.spinner {
  width: 28px;
  height: 28px;
  border: 3px solid var(--fluent-border);
  border-top-color: var(--fluent-accent);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.result-count {
  font-size: 12px;
  color: var(--fluent-text-secondary);
}

/* 歌曲列表 */
.result-toolbar {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 8px 4px 12px;
}

.play-all {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 36px;
  padding: 0 18px;
  border: none;
  border-radius: 18px;
  background: var(--fluent-accent);
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: filter 0.18s ease;
}

.play-all:hover {
  filter: brightness(1.1);
}

.play-all svg {
  width: 16px;
  height: 16px;
}

.switch-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  height: 36px;
  padding: 0 16px;
  border: 1px solid var(--fluent-border);
  border-radius: 18px;
  background: transparent;
  color: var(--fluent-text);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.18s ease, border-color 0.18s ease, opacity 0.18s ease;
}

.switch-btn:hover:not(:disabled) {
  background: var(--fluent-bg-hover);
  border-color: var(--fluent-accent);
}

.switch-btn:disabled {
  opacity: 0.5;
  cursor: default;
}

.switch-btn svg {
  width: 15px;
  height: 15px;
}

.song-list {
  display: flex;
  flex-direction: column;
}

.list-header,
.song-item {
  display: grid;
  grid-template-columns: 48px 32px 2fr 1.2fr 1.2fr 56px 132px;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
}

.list-header {
  position: sticky;
  top: 0;
  font-size: 11px;
  font-weight: 600;
  color: var(--fluent-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.4px;
  background: var(--fluent-bg-glass);
  backdrop-filter: blur(10px);
  z-index: 1;
}

.song-item {
  border-radius: 8px;
  cursor: default;
  transition: background 0.18s ease;
}

.song-item:hover {
  background: var(--fluent-bg-hover);
}

.song-item.playing {
  background: var(--fluent-accent);
  color: #fff;
}

.song-item.playing .secondary-text {
  color: rgba(255, 255, 255, 0.75);
}

.col-cover {
  position: relative;
  width: 44px;
  height: 44px;
  border-radius: 8px;
  overflow: hidden;
  background: var(--fluent-bg-active);
  flex-shrink: 0;
}

.cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.cover-fallback {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--fluent-text-secondary);
}

.cover-fallback svg {
  width: 20px;
  height: 20px;
}

.cover-play {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: rgba(0, 0, 0, 0.45);
  color: #fff;
  opacity: 0;
  cursor: pointer;
  transition: opacity 0.18s ease;
}

.col-cover:hover .cover-play {
  opacity: 1;
}

.cover-play svg {
  width: 20px;
  height: 20px;
}

.col-index {
  text-align: center;
  font-size: 12px;
  color: var(--fluent-text-secondary);
}

.col-title {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.primary-text {
  font-size: 13px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.secondary-text {
  font-size: 12px;
  color: var(--fluent-text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.col-artist,
.col-album,
.col-duration {
  min-width: 0;
}

.col-action {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 6px;
}

.source-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 10px;
  background: var(--fluent-bg-active);
  color: var(--fluent-text-secondary);
  white-space: nowrap;
}

.song-item.playing .source-badge {
  background: rgba(255, 255, 255, 0.2);
  color: #fff;
}

.icon-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 50%;
  background: transparent;
  color: var(--fluent-text-secondary);
  cursor: pointer;
  transition: background 0.18s ease, color 0.18s ease;
}

.icon-btn:hover {
  background: var(--fluent-bg-active);
  color: var(--fluent-text);
}

.icon-btn.active {
  color: #ff5c7c;
}

.icon-btn svg {
  width: 16px;
  height: 16px;
}

.toast {
  position: fixed;
  left: 50%;
  bottom: 86px;
  transform: translateX(-50%);
  padding: 8px 16px;
  border-radius: 18px;
  background: rgba(0, 0, 0, 0.78);
  color: #fff;
  font-size: 13px;
  z-index: 120;
  pointer-events: none;
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.35);
}

.toast-fade-enter-active,
.toast-fade-leave-active {
  transition: opacity 0.25s ease, transform 0.25s ease;
}

.toast-fade-enter-from,
.toast-fade-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(8px);
}
</style>
