<script lang="ts" setup>
import { ref, watch, computed } from 'vue'
import { OnlineComments } from '../../bindings/sugarplayer/app'
import type { OnlineSong, OnlineComment, OnlineCommentPage } from '../../bindings/sugarplayer/models'

const props = defineProps<{
  song: OnlineSong | null
}>()

const kind = ref<'latest' | 'hot'>('latest')
const page = ref(1)
const loading = ref(false)
const error = ref('')
const data = ref<OnlineCommentPage | null>(null)
const comments = ref<OnlineComment[]>([])

const sourceLabel: Record<string, string> = {
  wy: '网易云',
  tx: 'QQ',
  kg: '酷狗',
  kw: '酷我',
  mg: '咪咕',
}

const hasMore = computed(() => data.value != null && page.value < (data.value.maxPage || 1))

function formatTime(ms: number): string {
  if (!ms || ms <= 0) return ''
  const diff = Date.now() - ms
  const min = Math.floor(diff / 60000)
  if (min < 1) return '刚刚'
  if (min < 60) return `${min} 分钟前`
  const hour = Math.floor(min / 60)
  if (hour < 24) return `${hour} 小时前`
  const day = Math.floor(hour / 24)
  if (day < 30) return `${day} 天前`
  const d = new Date(ms)
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

async function load(reset = true) {
  if (!props.song) return
  if (reset) {
    page.value = 1
    comments.value = []
  }
  loading.value = true
  error.value = ''
  try {
    const res = await OnlineComments(props.song, kind.value, page.value)
    data.value = res
    if (reset) comments.value = res.comments
    else comments.value = comments.value.concat(res.comments)
  } catch (e) {
    error.value = e instanceof Error ? e.message : '评论加载失败'
  } finally {
    loading.value = false
  }
}

function loadMore() {
  if (loading.value || !hasMore.value) return
  page.value++
  load(false)
}

function switchKind(k: 'latest' | 'hot') {
  if (k === kind.value) return
  kind.value = k
  load(true)
}

watch(
  () => props.song,
  (s) => {
    if (s) load(true)
    else {
      comments.value = []
      data.value = null
    }
  },
  { immediate: true }
)

watch(kind, () => load(true))
</script>

<template>
  <div class="comment-list">
    <div v-if="!song" class="comment-empty">暂无可显示的评论</div>
    <template v-else>
      <div class="comment-head">
        <div class="seg">
          <button :class="['seg-item', { active: kind === 'latest' }]" @click="switchKind('latest')">最新</button>
          <button :class="['seg-item', { active: kind === 'hot' }]" @click="switchKind('hot')">热门</button>
        </div>
        <span class="comment-total" v-if="data">共 {{ data.total }} 条 · {{ sourceLabel[song.source] || song.source }}</span>
      </div>

      <div v-if="loading && comments.length === 0" class="comment-state">
        <div class="spinner"></div>
        <span>加载中…</span>
      </div>
      <div v-else-if="error" class="comment-state error"><span>{{ error }}</span></div>
      <div v-else-if="comments.length === 0" class="comment-state"><span>还没有评论</span></div>

      <div v-else class="comment-items">
        <div v-for="c in comments" :key="c.id" class="comment-item">
          <img v-if="c.avatar" :src="c.avatar" class="c-avatar" alt="" loading="lazy" />
          <div v-else class="c-avatar fallback">{{ (c.userName || '?').slice(0, 1) }}</div>
          <div class="c-body">
            <div class="c-meta">
              <span class="c-name">{{ c.userName || '匿名' }}</span>
              <span v-if="c.location" class="c-loc">{{ c.location }}</span>
              <span class="c-time">{{ formatTime(c.time) }}</span>
            </div>
            <div class="c-text">{{ c.text }}</div>
            <div v-if="c.images && c.images.length" class="c-images">
              <img v-for="(img, i) in c.images" :key="i" :src="img" class="c-image" alt="" loading="lazy" />
            </div>
            <div class="c-foot">
              <span class="c-like">
                <svg class="like-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/></svg>
                {{ c.likedCount || 0 }}
              </span>
              <span v-if="c.replyNum" class="c-reply-num">{{ c.replyNum }} 条回复</span>
            </div>
            <div v-if="c.reply && c.reply.length" class="c-replies">
              <div v-for="r in c.reply" :key="r.id" class="c-reply">
                <img v-if="r.avatar" :src="r.avatar" class="r-avatar" alt="" loading="lazy" />
                <div v-else class="r-avatar fallback">{{ (r.userName || '?').slice(0, 1) }}</div>
                <div class="r-body">
                  <div class="c-meta">
                    <span class="c-name">{{ r.userName || '匿名' }}</span>
                    <span class="c-time">{{ formatTime(r.time) }}</span>
                  </div>
                  <div class="c-text">{{ r.text }}</div>
                  <div class="c-foot"><span class="c-like">
                    <svg class="like-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/></svg>
                    {{ r.likedCount || 0 }}
                  </span></div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <button v-if="hasMore" class="load-more" :disabled="loading" @click="loadMore">
          {{ loading ? '加载中…' : '加载更多' }}
        </button>
      </div>
    </template>
  </div>
</template>

<style scoped>
.comment-list {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
}

.comment-empty,
.comment-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 40px 20px;
  color: var(--fluent-text-secondary);
  font-size: 13px;
}

.comment-state.error {
  color: #ff8080;
}

.spinner {
  width: 24px;
  height: 24px;
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

.comment-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 4px 4px 10px;
  flex-shrink: 0;
}

.seg {
  display: flex;
  gap: 2px;
  padding: 3px;
  border-radius: 16px;
  background: var(--fluent-bg-active);
}

.seg-item {
  padding: 4px 14px;
  border: none;
  border-radius: 13px;
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
}

.comment-total {
  font-size: 12px;
  color: var(--fluent-text-secondary);
}

.comment-items {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
  padding-right: 4px;
}

.comment-item {
  display: flex;
  gap: 10px;
  padding: 12px 4px;
  border-bottom: 1px solid var(--fluent-border);
}

.c-avatar,
.r-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  object-fit: cover;
  flex-shrink: 0;
  background: var(--fluent-bg-active);
}

.c-avatar.fallback,
.r-avatar.fallback {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--fluent-text-secondary);
  font-size: 15px;
  font-weight: 600;
}

.r-avatar {
  width: 28px;
  height: 28px;
  font-size: 12px;
}

.c-body {
  flex: 1;
  min-width: 0;
}

.c-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.c-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--fluent-text);
}

.c-loc,
.c-time {
  font-size: 11px;
  color: var(--fluent-text-secondary);
}

.c-text {
  margin: 4px 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--fluent-text);
  white-space: pre-wrap;
  word-break: break-word;
}

.c-images {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin: 6px 0;
}

.c-image {
  width: 88px;
  height: 88px;
  border-radius: 8px;
  object-fit: cover;
  cursor: pointer;
}

.c-foot {
  display: flex;
  align-items: center;
  gap: 14px;
  font-size: 12px;
  color: var(--fluent-text-secondary);
}

.c-like {
  display: inline-flex;
  align-items: center;
  gap: 3px;
}

.like-icon {
  width: 13px;
  height: 13px;
  fill: currentColor;
  flex-shrink: 0;
}

.c-replies {
  margin-top: 10px;
  padding: 8px 10px;
  border-radius: 10px;
  background: var(--fluent-bg-active);
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.c-reply {
  display: flex;
  gap: 8px;
}

.r-body {
  flex: 1;
  min-width: 0;
}

.load-more {
  display: block;
  width: 100%;
  margin: 14px 0 8px;
  padding: 9px;
  border: 1px solid var(--fluent-border);
  border-radius: 16px;
  background: transparent;
  color: var(--fluent-text);
  font-size: 13px;
  cursor: pointer;
  transition: background 0.18s ease;
}

.load-more:hover:not(:disabled) {
  background: var(--fluent-bg-hover);
}

.load-more:disabled {
  opacity: 0.5;
  cursor: default;
}
</style>
