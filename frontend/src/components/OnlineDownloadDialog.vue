<script lang="ts" setup>
import { ref, watch, computed, inject, type Ref } from 'vue'
import {
  OnlineIsUnlocked,
  OnlineVerifyKey,
  OnlineDownload,
  OpenMusicFolder,
  OpenInExplorer,
} from '../../bindings/sugarplayer/app'
import type { OnlineSong, OnlineDownloadResult } from '../../bindings/sugarplayer/models'
import { currentOnlineSong } from '../composables/onlineState'
import type { AppSettings } from '../composables/useConfig'

const props = defineProps<{
  show: boolean
  song: OnlineSong | null
}>()

const settingsRef = inject<Ref<AppSettings>>('settings')
const settings = computed(() => settingsRef?.value)

const emit = defineEmits<{
  (e: 'close'): void
}>()

const unlocked = ref(false)
const checking = ref(false)
const keyInput = ref('')
const keyError = ref('')

const dir = ref('')
const withLyrics = ref(true)
const withCover = ref(true)
const embed = ref(true)

const downloading = ref(false)
const error = ref('')
const result = ref<OnlineDownloadResult | null>(null)

const dirHint = computed(() => dir.value || '默认：音乐库 / SugarPlayer')

// 下载使用当前音质：若下载的正是正在播放的歌曲，则沿用其已选音质；否则取歌曲自身 extra 中的音质
function parseQuality(extra?: string): string {
  if (!extra) return ''
  try {
    const obj = JSON.parse(extra) as Record<string, string>
    return obj['quality'] || ''
  } catch {
    return ''
  }
}

const currentQuality = computed(() => {
  const song = props.song
  if (!song) return settings.value?.quality || 'standard'
  const cur = currentOnlineSong.value
  if (cur && cur.source === song.source && cur.id === song.id) {
    const q = parseQuality(cur.extra)
    if (q) return q
  }
  const songQuality = parseQuality(song.extra)
  if (songQuality) return songQuality
  // 如果没有设置过音质，使用用户设置的默认音质
  return settings.value?.quality || 'standard'
})

watch(
  () => [props.show, props.song],
  async ([show]) => {
    if (show) {
      error.value = ''
      result.value = null
      keyInput.value = ''
      keyError.value = ''
      downloading.value = false
      try {
        unlocked.value = await OnlineIsUnlocked()
      } catch {
        unlocked.value = false
      }
    }
  },
  { immediate: true },
)

async function chooseDir() {
  try {
    const folder = await OpenMusicFolder()
    if (folder) dir.value = folder
  } catch {
    // 用户取消
  }
}

async function startDownload() {
  if (!props.song) return
  if (!unlocked.value) {
    checking.value = true
    keyError.value = ''
    try {
      const ok = await OnlineVerifyKey(keyInput.value)
      if (!ok) {
        keyError.value = '密钥不正确'
        checking.value = false
        return
      }
      unlocked.value = true
    } catch {
      keyError.value = '密钥校验失败'
      checking.value = false
      return
    }
    checking.value = false
  }

  downloading.value = true
  error.value = ''
  result.value = null
  try {
    const res = await OnlineDownload(props.song, {
      dir: dir.value,
      withLyrics: withLyrics.value,
      withCover: withCover.value,
      embed: embed.value,
      quality: currentQuality.value,
    })
    result.value = res
  } catch (e) {
    error.value = e instanceof Error ? e.message : '下载失败'
  } finally {
    downloading.value = false
  }
}

function openFolder() {
  const target = result.value?.path || dir.value
  if (target) OpenInExplorer(target).catch(() => {})
}
</script>

<template>
  <div v-if="show" class="modal-mask" @click.self="emit('close')">
    <div class="dialog">
      <div class="dialog-head">
        <span class="dialog-title">下载歌曲</span>
        <button class="close-btn" @click="emit('close')">×</button>
      </div>

      <div v-if="song" class="song-info">
        <span class="song-name">{{ song.name }}</span>
        <span class="song-artist">{{ song.artist }}</span>
      </div>

      <div v-if="!unlocked" class="key-block">
        <label class="field-label">下载密钥（仅首次需要，仅供个人使用）</label>
        <input
          v-model="keyInput"
          class="key-input"
          type="password"
          placeholder="请输入下载密钥"
          @keydown.enter="startDownload"
        />
        <span v-if="keyError" class="key-error">{{ keyError }}</span>
      </div>

      <div class="field">
        <label class="field-label">保存位置</label>
        <div class="dir-row">
          <span class="dir-text" :title="dir">{{ dirHint }}</span>
          <button class="link-btn" @click="chooseDir">选择文件夹</button>
        </div>
      </div>

      <div class="options">
        <label class="opt">
          <input v-model="withLyrics" type="checkbox" />
          <span>下载歌词（.lrc）</span>
        </label>
        <label class="opt">
          <input v-model="withCover" type="checkbox" />
          <span>下载封面</span>
        </label>
        <label class="opt">
          <input v-model="embed" type="checkbox" />
          <span>内嵌到音频（封面 + 歌词写入文件）</span>
        </label>
      </div>

      <div v-if="error" class="result-error">{{ error }}</div>

      <div v-if="result" class="result">
        <div class="result-ok">下载完成</div>
        <div class="result-line">
          <span class="result-key">音频</span>
          <span class="result-val" :title="result.path">{{ result.path }}</span>
        </div>
        <div v-if="result.lyricPath" class="result-line">
          <span class="result-key">歌词</span>
          <span class="result-val" :title="result.lyricPath">{{ result.lyricPath }}</span>
        </div>
        <div v-if="result.coverPath" class="result-line">
          <span class="result-key">封面</span>
          <span class="result-val" :title="result.coverPath">{{ result.coverPath }}</span>
        </div>
        <div v-if="result.warning" class="result-warn">{{ result.warning }}</div>
      </div>

      <div class="dialog-foot">
        <button v-if="result" class="ghost-btn" @click="openFolder">打开文件夹</button>
        <button class="ghost-btn" @click="emit('close')">关闭</button>
        <button
          v-if="!result"
          class="primary-btn"
          :disabled="downloading || (!unlocked && checking)"
          @click="startDownload"
        >
          {{ downloading ? '下载中…' : '下载' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  backdrop-filter: blur(4px);
}

.dialog {
  width: 420px;
  max-width: 92vw;
  background: var(--fluent-bg-glass);
  border: 1px solid var(--fluent-border);
  border-radius: 12px;
  padding: 18px 20px 16px;
  box-shadow: 0 18px 50px rgba(0, 0, 0, 0.4);
  color: var(--fluent-text);
}

.dialog-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.dialog-title {
  font-size: 15px;
  font-weight: 600;
}

.close-btn {
  border: none;
  background: transparent;
  color: var(--fluent-text-secondary);
  font-size: 20px;
  line-height: 1;
  cursor: pointer;
}

.song-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 8px 10px;
  background: var(--fluent-bg-hover);
  border-radius: 8px;
  margin-bottom: 12px;
}

.song-name {
  font-size: 13px;
  font-weight: 600;
}

.song-artist {
  font-size: 12px;
  color: var(--fluent-text-secondary);
}

.field {
  margin-bottom: 12px;
}

.field-label {
  display: block;
  font-size: 12px;
  color: var(--fluent-text-secondary);
  margin-bottom: 6px;
}

.key-block {
  margin-bottom: 12px;
}

.key-input {
  width: 100%;
  height: 34px;
  padding: 0 10px;
  border-radius: 8px;
  border: 1px solid var(--fluent-input-border);
  background: var(--fluent-input-bg);
  color: var(--fluent-text);
  font-size: 13px;
  outline: none;
}

.key-input:focus {
  border-color: var(--fluent-accent);
}

.key-error {
  display: block;
  margin-top: 6px;
  font-size: 12px;
  color: #ff8080;
}

.dir-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.dir-text {
  flex: 1;
  font-size: 12px;
  color: var(--fluent-text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.link-btn {
  border: none;
  background: transparent;
  color: var(--fluent-accent);
  font-size: 12px;
  cursor: pointer;
  white-space: nowrap;
}

.link-btn:hover {
  text-decoration: underline;
}

.options {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 8px;
}

.opt {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  cursor: pointer;
}

.opt input {
  accent-color: var(--fluent-accent);
}

.result-error {
  margin: 8px 0;
  font-size: 12px;
  color: #ff8080;
}

.result {
  margin: 10px 0;
  padding: 10px 12px;
  background: var(--fluent-bg-hover);
  border-radius: 8px;
}

.result-ok {
  font-size: 13px;
  font-weight: 600;
  color: #6fd08c;
  margin-bottom: 8px;
}

.result-line {
  display: flex;
  gap: 8px;
  font-size: 12px;
  margin-bottom: 4px;
}

.result-key {
  color: var(--fluent-text-secondary);
  flex-shrink: 0;
  width: 36px;
}

.result-val {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.result-warn {
  margin-top: 6px;
  font-size: 12px;
  color: #e0b050;
}

.dialog-foot {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 12px;
}

.primary-btn {
  height: 34px;
  padding: 0 18px;
  border: none;
  border-radius: 17px;
  background: var(--fluent-accent);
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}

.primary-btn:disabled {
  opacity: 0.5;
  cursor: default;
}

.ghost-btn {
  height: 34px;
  padding: 0 16px;
  border: 1px solid var(--fluent-border);
  border-radius: 17px;
  background: transparent;
  color: var(--fluent-text);
  font-size: 13px;
  cursor: pointer;
}

.ghost-btn:hover {
  background: var(--fluent-bg-hover);
}
</style>