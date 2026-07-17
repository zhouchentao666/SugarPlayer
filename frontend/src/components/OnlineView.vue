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
  OnlinePlaylistCategories,
  OnlineCategoryPlaylists,
} from '../../bindings/sugarplayer/app'
import type { OnlineSong, OnlineSource, OnlineCollection, OnlineCategorySource, OnlineCategoryItem } from '../../bindings/sugarplayer/models'
import type { Song } from '../types'
import OnlineDownloadDialog from './OnlineDownloadDialog.vue'
import CommentList from './CommentList.vue'

const props = defineProps<{
  currentSong: Song | null
  autoSwitch: boolean
  pinnedCollections: OnlineCollection[]
  openCollection: OnlineCollection | null
  favoritedKeys: string[]
  section: 'search' | 'discover'
  searchSources: string[]
  searchHistory: string[]
}>()

const emit = defineEmits<{
  (e: 'play', list: OnlineSong[], index: number): void
  (e: 'add-to-queue', song: OnlineSong): void
  (e: 'request-add-to-playlist', song: OnlineSong): void
  (e: 'unfavorite', song: OnlineSong): void
  (e: 'toggle-pin', collection: OnlineCollection): void
  (e: 'update:searchSources', sources: string[]): void
  (e: 'update:searchHistory', history: string[]): void
}>()

const downloadSong = ref<OnlineSong | null>(null)
const showDownload = ref(false)
const showCommentsSong = ref<OnlineSong | null>(null)
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
  return props.favoritedKeys.includes(favoriteKey(song))
}

function toggleFavorite(song: OnlineSong) {
  if (isFavorited(song)) {
    emit('unfavorite', song)
    showToast('已取消收藏')
  } else {
    emit('request-add-to-playlist', song)
  }
}

function openDownload(song: OnlineSong) {
  downloadSong.value = song
  showDownload.value = true
}

function openComments(song: OnlineSong) {
  showCommentsSong.value = song
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

// 发现页：平台单选 + 分类标签（推荐 / 我的 / 分类）
const discoverPlatform = ref('') // '' 表示全部
const discoverCategory = ref<'recommend' | 'user' | 'category' | null>(null)

const showRecommendTab = computed(() => {
  if (discoverPlatform.value === '') return allSources.value.some(s => s.recommend)
  const s = allSources.value.find(x => x.id === discoverPlatform.value)
  return !!s?.recommend
})

const showUserTab = computed(() => {
  if (discoverPlatform.value === '') return allSources.value.some(s => s.userPlaylists)
  const s = allSources.value.find(x => x.id === discoverPlatform.value)
  return !!s?.userPlaylists
})

// 可用的分类标签（顺序即默认顺序，第一个为默认）
const discoverTabs = computed(() => {
  const tabs: { key: 'recommend' | 'user' | 'category'; label: string }[] = []
  if (showRecommendTab.value) tabs.push({ key: 'recommend', label: '推荐' })
  if (showUserTab.value) tabs.push({ key: 'user', label: '我的' })
  tabs.push({ key: 'category', label: '分类' })
  return tabs
})

const defaultDiscoverTab = computed<'recommend' | 'user' | 'category'>(() => discoverTabs.value[0]?.key ?? 'category')

// 歌单分类树（按平台筛选）
const categoryTree = ref<OnlineCategorySource[]>([])
const categoryLoading = ref(false)
const categoryError = ref('')
const expandedGroups = ref<Record<string, boolean>>({})
const activeCategory = ref<{ source: string; id: string; name: string } | null>(null)

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
    // 优先使用已保存的勾选；首次启动默认全部启用并写回持久化
    if (props.searchSources && props.searchSources.length) {
      selected.value = props.searchSources.filter(id => sources.some(s => s.id === id))
    } else {
      selected.value = sources.filter(s => s.enabled).map(s => s.id)
      emit('update:searchSources', [...selected.value])
    }
    if (props.section === 'discover') ensureDiscoverCategory()
  } catch {
    allSources.value = []
  }
})

// 切到「发现」时默认选中第一个分类标签
watch(
  () => props.section,
  (s) => {
    if (s === 'discover') ensureDiscoverCategory()
  }
)

function toggleSource(id: string) {
  const i = selected.value.indexOf(id)
  if (i >= 0) selected.value.splice(i, 1)
  else selected.value.push(id)
  emit('update:searchSources', [...selected.value])
}

function selectAll() {
  selected.value = allSources.value.map(s => s.id)
  emit('update:searchSources', [...selected.value])
}

function clearAll() {
  selected.value = []
  emit('update:searchSources', [...selected.value])
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
  // 记录搜索历史（去重、最多保留 20 条）
  const hist = [q, ...props.searchHistory.filter(h => h !== q)].slice(0, 20)
  emit('update:searchHistory', hist)
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

// 点击历史记录：回填并重新搜索
function applyHistory(h: string) {
  keyword.value = h
  search()
}

function removeHistory(h: string) {
  emit('update:searchHistory', props.searchHistory.filter(x => x !== h))
}

function clearHistory() {
  emit('update:searchHistory', [])
}

// ---------- 发现页：平台单选 + 分类（推荐 / 我的 / 分类） ----------
function selectDiscoverPlatform(id: string) {
  discoverPlatform.value = id
  // 当前分类在新平台不可用时重置
  if (discoverCategory.value === 'user' && !showUserTab.value) discoverCategory.value = null
  if (discoverCategory.value === 'recommend' && !showRecommendTab.value) discoverCategory.value = null
  // 已选分类则按新平台重新加载（分类标签需重建树）
  if (discoverCategory.value) applyDiscoverCategory()
}

function selectDiscoverCategory(cat: 'recommend' | 'user' | 'category') {
  if (cat === 'user' && !showUserTab.value) return
  if (cat === 'recommend' && !showRecommendTab.value) return
  discoverCategory.value = cat
  applyDiscoverCategory()
}

// 确保发现页有默认选中的标签（默认第一个）
function ensureDiscoverCategory() {
  if (!discoverCategory.value || !discoverTabs.value.some(t => t.key === discoverCategory.value)) {
    discoverCategory.value = defaultDiscoverTab.value
  }
  applyDiscoverCategory()
}

// 进入某分类标签：推荐/我的直接拉取；分类则加载树（默认展开第一个分组）
async function applyDiscoverCategory() {
  const cat = discoverCategory.value
  if (!cat) return
  activeCategory.value = null
  if (cat === 'recommend' || cat === 'user') {
    await loadDiscover(cat)
    return
  }
  await loadCategoryTree()
}

async function loadDiscover(cat: 'recommend' | 'user') {
  loading.value = true
  error.value = ''
  loaded.value = false
  mode.value = 'list'
  listKind.value = 'collections'
  currentCollection.value = null
  discoverOrigin.value = true
  activeCategory.value = null
  const src = discoverPlatform.value ? [discoverPlatform.value] : []
  try {
    if (cat === 'recommend') {
      listTitle.value = '每日推荐歌单'
      collections.value = await OnlineRecommendPlaylists(src)
    } else {
      listTitle.value = '我的歌单'
      collections.value = await OnlineUserPlaylists(src)
    }
    loaded.value = true
  } catch (e) {
    collections.value = []
    error.value = e instanceof Error ? e.message : '加载失败'
  } finally {
    loading.value = false
  }
}

// 歌单分类树
async function loadCategoryTree() {
  mode.value = 'discover' // 关键：切到分类时复位视图，避免仍显示上一标签（推荐/我的）的歌单
  collections.value = []
  loaded.value = false
  categoryLoading.value = true
  categoryError.value = ''
  try {
    const src = discoverPlatform.value ? [discoverPlatform.value] : []
    categoryTree.value = await OnlinePlaylistCategories(src)
    // 默认展开第一个分组
    const firstKey = firstGroupKey()
    expandedGroups.value = firstKey ? { [firstKey]: true } : {}
  } catch (e) {
    categoryTree.value = []
    categoryError.value = e instanceof Error ? e.message : '分类加载失败'
  } finally {
    categoryLoading.value = false
  }
}

function firstGroupKey(): string {
  for (const src of categoryTree.value) {
    for (const g of src.groups) {
      return `${src.source}|${g.name}`
    }
  }
  return ''
}

function isGroupExpanded(source: string, name: string): boolean {
  return expandedGroups.value[`${source}|${name}`] ?? false
}

function toggleGroup(source: string, name: string) {
  const key = `${source}|${name}`
  expandedGroups.value = { ...expandedGroups.value, [key]: !isGroupExpanded(source, name) }
}

// 点击分类：直接在主界面展示该分类的歌单（不打开新界面）
async function selectCategoryItem(item: OnlineCategoryItem) {
  activeCategory.value = { source: item.source, id: item.id, name: item.name }
  loading.value = true
  error.value = ''
  loaded.value = false
  mode.value = 'list'
  listKind.value = 'collections'
  currentCollection.value = null
  discoverOrigin.value = true
  try {
    listTitle.value = `${sourceNameMap.value[item.source] ?? item.source} · ${item.name}`
    collections.value = await OnlineCategoryPlaylists(item.source, item.id, item.name)
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

// 返回按钮：推荐 / 我的 视图下不显示（标签常驻，无需返回）
const showListBack = computed(() => {
  if (discoverOrigin.value && (discoverCategory.value === 'recommend' || discoverCategory.value === 'user')) {
    return false
  }
  return true
})

function goBack() {
  if (currentCollection.value) {
    currentCollection.value = null
    listKind.value = 'collections'
  } else if (activeCategory.value) {
    // 从分类歌单返回分类树
    activeCategory.value = null
    mode.value = 'discover'
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

// 单击在线歌曲：添加到播放列表（不替换）
function addToQueue(song: OnlineSong) {
  emit('add-to-queue', song)
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
    <div v-if="section === 'search'" class="search-panel">
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

    <!-- 发现页：平台单选 + 分类（推荐 / 我的 / 分类） -->
    <div v-if="section === 'discover' && !currentCollection" class="discover-panel">
      <div class="platform-bar">
        <button
          :class="['platform-chip', { active: discoverPlatform === '' }]"
          @click="selectDiscoverPlatform('')"
        >
          全部
        </button>
        <button
          v-for="src in allSources"
          :key="src.id"
          :class="['platform-chip', { active: discoverPlatform === src.id }]"
          @click="selectDiscoverPlatform(src.id)"
        >
          {{ src.name }}
        </button>
      </div>
      <div class="cat-tabs">
        <button
          v-for="tab in discoverTabs"
          :key="tab.key"
          :class="['cat-tab', { active: discoverCategory === tab.key }]"
          @click="selectDiscoverCategory(tab.key)"
        >
          {{ tab.label }}
        </button>
      </div>
    </div>

    <div class="results">
      <!-- 搜索历史（仅搜索区） -->
      <div v-if="section === 'search' && mode === 'discover'" class="discover">
        <h2 class="discover-title">搜索历史</h2>
        <div v-if="searchHistory.length" class="history-list">
          <div
            v-for="h in searchHistory"
            :key="h"
            class="history-item"
            @click="applyHistory(h)"
          >
            <svg class="history-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="11" cy="11" r="7" />
              <line x1="21" y1="21" x2="16.65" y2="16.65" />
            </svg>
            <span class="history-text">{{ h }}</span>
            <button
              class="history-del"
              title="删除"
              @click.stop="removeHistory(h)"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                <line x1="6" y1="6" x2="18" y2="18" />
                <line x1="18" y1="6" x2="6" y2="18" />
              </svg>
            </button>
          </div>
          <button class="history-clear" @click="clearHistory">清空历史</button>
        </div>
        <p v-else class="discover-hint">还没有搜索记录，搜点什么吧～</p>
      </div>

      <!-- 发现页：歌单分类树（可展开 / 折叠） -->
      <div
        v-else-if="section === 'discover' && mode === 'discover' && discoverCategory === 'category' && !activeCategory"
        class="discover category-tree"
      >
        <div v-if="categoryLoading" class="state">
          <div class="spinner"></div>
          <span>正在加载分类…</span>
        </div>
        <div v-else-if="categoryError" class="state error">
          <span>{{ categoryError }}</span>
        </div>
        <div v-else-if="!categoryTree.length" class="state">
          <span>该音源暂不支持歌单分类</span>
        </div>
        <template v-else>
          <div v-for="src in categoryTree" :key="src.source" class="cat-source">
            <div class="cat-source-title">{{ src.name }}</div>
            <div v-for="group in src.groups" :key="group.name" class="cat-group">
              <button class="cat-group-head" @click="toggleGroup(src.source, group.name)">
                <svg
                  class="cat-arrow"
                  :class="{ open: isGroupExpanded(src.source, group.name) }"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <polyline points="9 6 15 12 9 18" />
                </svg>
                <span>{{ group.name }}</span>
              </button>
              <div v-show="isGroupExpanded(src.source, group.name)" class="cat-chip-list">
                <button
                  v-for="item in group.categories"
                  :key="item.id + item.name"
                  :class="['cat-chip', { hot: item.hot }]"
                  @click="selectCategoryItem(item)"
                >
                  {{ item.name }}
                  <span v-if="item.hot" class="cat-hot">HOT</span>
                </button>
              </div>
            </div>
          </div>
        </template>
      </div>

      <!-- 发现页：未选分类时的提示 -->
      <div v-else-if="section === 'discover' && mode === 'discover'" class="discover">
        <p class="discover-hint">选择上方分类查看对应歌单</p>
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
          <button v-if="showListBack" class="back-btn" @click="goBack">
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
              @click="addToQueue(song)"
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
                <button class="icon-btn" title="评论" @click.stop="openComments(song)">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z" />
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

      <Transition name="modal-fade">
        <div v-if="showCommentsSong" class="comment-modal-mask" @click.self="showCommentsSong = null">
          <div class="comment-modal">
            <div class="cm-head">
              <div class="cm-title">
                <span>歌曲评论</span>
                <span v-if="showCommentsSong" class="cm-song">{{ showCommentsSong.name }} - {{ showCommentsSong.artist }}</span>
              </div>
              <button class="cm-close" title="关闭" @click="showCommentsSong = null">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                  <line x1="6" y1="6" x2="18" y2="18" />
                  <line x1="18" y1="6" x2="6" y2="18" />
                </svg>
              </button>
            </div>
            <div class="cm-body">
              <CommentList :song="showCommentsSong" />
            </div>
          </div>
        </div>
      </Transition>

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

/* 发现页：平台单选 + 分类 */
.discover-panel {
  padding: 16px 24px 4px;
  flex-shrink: 0;
}

.platform-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.platform-chip {
  padding: 6px 14px;
  border: 1px solid var(--fluent-border);
  border-radius: 16px;
  background: transparent;
  color: var(--fluent-text-secondary);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.18s ease;
}

.platform-chip:hover {
  background: var(--fluent-bg-hover);
  color: var(--fluent-text);
}

.platform-chip.active {
  background: var(--fluent-accent);
  border-color: var(--fluent-accent);
  color: #fff;
}

.cat-tabs {
  display: flex;
  gap: 4px;
  margin-top: 14px;
  border-bottom: 1px solid var(--fluent-border);
}

.cat-tab {
  padding: 8px 16px;
  border: none;
  border-bottom: 2px solid transparent;
  background: transparent;
  color: var(--fluent-text-secondary);
  font-size: 14px;
  cursor: pointer;
  transition: color 0.18s ease, border-color 0.18s ease;
}

.cat-tab:hover {
  color: var(--fluent-text);
}

.cat-tab.active {
  color: var(--fluent-accent);
  border-bottom-color: var(--fluent-accent);
  font-weight: 600;
}

/* 歌单分类树 */
.category-tree {
  padding-top: 8px;
}

.cat-source {
  margin-bottom: 18px;
}

.cat-source-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--fluent-text-secondary);
  margin-bottom: 8px;
}

.cat-group {
  margin-bottom: 6px;
}

.cat-group-head {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 100%;
  padding: 8px 4px;
  border: none;
  background: transparent;
  color: var(--fluent-text);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  text-align: left;
  transition: color 0.18s ease;
}

.cat-group-head:hover {
  color: var(--fluent-accent);
}

.cat-arrow {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
  transition: transform 0.18s ease;
}

.cat-arrow.open {
  transform: rotate(90deg);
}

.cat-chip-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 4px 0 8px 20px;
}

.cat-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 5px 12px;
  border: 1px solid var(--fluent-border);
  border-radius: 14px;
  background: transparent;
  color: var(--fluent-text);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.18s ease;
}

.cat-chip:hover {
  background: var(--fluent-bg-hover);
  border-color: var(--fluent-accent);
}

.cat-chip.hot {
  border-color: rgba(255, 92, 124, 0.5);
  color: #ff5c7c;
}

.cat-hot {
  font-size: 9px;
  font-weight: 700;
  padding: 0 3px;
  border-radius: 4px;
  background: #ff5c7c;
  color: #fff;
  line-height: 14px;
}

.discover-hint {
  margin: 18px 0 0;
  font-size: 12px;
  color: var(--fluent-text-secondary);
}

/* 搜索历史 */
.history-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-width: 520px;
}

.history-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--fluent-border);
  border-radius: 10px;
  background: var(--fluent-bg-glass);
  color: var(--fluent-text);
  font-size: 13px;
  text-align: left;
  cursor: pointer;
  transition: background 0.18s ease, border-color 0.18s ease;
}

.history-item:hover {
  background: var(--fluent-bg-hover);
  border-color: var(--fluent-accent);
}

.history-icon {
  width: 16px;
  height: 16px;
  color: var(--fluent-text-secondary);
  flex-shrink: 0;
}

.history-text {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.history-del {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 50%;
  background: transparent;
  color: var(--fluent-text-secondary);
  cursor: pointer;
  opacity: 0;
  flex-shrink: 0;
  transition: opacity 0.18s ease, background 0.18s ease, color 0.18s ease;
}

.history-item:hover .history-del {
  opacity: 1;
}

.history-del:hover {
  background: var(--fluent-bg-active);
  color: #ff8080;
}

.history-del svg {
  width: 14px;
  height: 14px;
}

.history-clear {
  align-self: flex-start;
  margin-top: 4px;
  padding: 6px 14px;
  border: 1px solid var(--fluent-border);
  border-radius: 14px;
  background: transparent;
  color: var(--fluent-text-secondary);
  font-size: 12px;
  cursor: pointer;
  transition: background 0.18s ease, color 0.18s ease;
}

.history-clear:hover {
  background: var(--fluent-bg-hover);
  color: var(--fluent-text);
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

.comment-modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 130;
}

.comment-modal {
  width: min(560px, 92vw);
  height: min(680px, 82vh);
  display: flex;
  flex-direction: column;
  background: var(--fluent-bg-glass);
  backdrop-filter: blur(40px) saturate(125%);
  -webkit-backdrop-filter: blur(40px) saturate(125%);
  border: 1px solid var(--fluent-border);
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.4);
  overflow: hidden;
}

.cm-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px;
  border-bottom: 1px solid var(--fluent-border);
  flex-shrink: 0;
}

.cm-title {
  display: flex;
  align-items: baseline;
  gap: 10px;
  min-width: 0;
}

.cm-title > span:first-child {
  font-size: 15px;
  font-weight: 700;
  color: var(--fluent-text);
}

.cm-song {
  font-size: 12px;
  color: var(--fluent-text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cm-close {
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 50%;
  background: transparent;
  color: var(--fluent-text-secondary);
  cursor: pointer;
  flex-shrink: 0;
  transition: background 0.18s ease, color 0.18s ease;
}

.cm-close:hover {
  background: var(--fluent-bg-hover);
  color: var(--fluent-text);
}

.cm-close svg {
  width: 18px;
  height: 18px;
}

.cm-body {
  flex: 1;
  min-height: 0;
  padding: 0 14px;
}

.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.2s ease;
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}
</style>
