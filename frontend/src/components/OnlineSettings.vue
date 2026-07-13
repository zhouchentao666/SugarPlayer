<script lang="ts" setup>
import { ref, onMounted, onUnmounted } from 'vue'
import {
  GetPlatformCookies,
  SetPlatformCookies,
  OnlineSources,
  QRLoginSources,
  CreateQRLogin,
  CheckQRLogin,
} from '../../bindings/sugarplayer/app'
import type { OnlineSource } from '../../bindings/sugarplayer/models'
import type { AppSettings } from '../composables/useConfig'
import ToggleSwitch from './settings/ToggleSwitch.vue'

const props = defineProps<{
  settings: AppSettings
}>()

const emit = defineEmits<{
  (e: 'update:settings', settings: AppSettings): void
  (e: 'close'): void
}>()

const platforms = ref<OnlineSource[]>([])
// 每个平台的 Cookie 文本（本地编辑态）
const cookies = ref<Record<string, string>>({})
const qrSources = ref<string[]>([])
const savedFlag = ref(false)
let savedTimer: ReturnType<typeof setTimeout> | null = null
let debounceTimer: ReturnType<typeof setTimeout> | null = null

function statusOf(id: string): 'set' | 'empty' {
  return (cookies.value[id] || '').trim() ? 'set' : 'empty'
}

onMounted(async () => {
  try {
    platforms.value = await OnlineSources()
  } catch {
    platforms.value = []
  }
  try {
    qrSources.value = await QRLoginSources()
  } catch {
    qrSources.value = []
  }
  try {
    const remote = await GetPlatformCookies()
    cookies.value = { ...remote }
  } catch {
    cookies.value = { ...(props.settings.platformCookies || {}) }
  }
})

onUnmounted(() => {
  if (savedTimer) clearTimeout(savedTimer)
  if (debounceTimer) clearTimeout(debounceTimer)
  stopQRPoll()
})

function onAutoSwitchChange(value: boolean) {
  emit('update:settings', { ...props.settings, autoSwitchInvalidSource: value })
}

// 立即保存 Cookie（注入 Go core.CM + 写入配置），并显示「已自动保存」提示。
async function persistCookies() {
  const clean: Record<string, string> = {}
  for (const [k, v] of Object.entries(cookies.value)) {
    const t = (v || '').trim()
    if (t) clean[k] = t
  }
  try {
    await SetPlatformCookies(clean)
  } catch {
    // 忽略 Go 端写入失败，仍更新前端状态
  }
  emit('update:settings', { ...props.settings, platformCookies: clean })
  savedFlag.value = true
  if (savedTimer) clearTimeout(savedTimer)
  savedTimer = setTimeout(() => (savedFlag.value = false), 1600)
}

// 输入 Cookie 后防抖 600ms 自动保存。
function onCookieInput() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    persistCookies()
  }, 600)
}

// ---------- 扫码登录 ----------
interface QRState {
  open: boolean
  source: string
  label: string
  image: string
  status: string
  loading: boolean
  error: string
}

const qr = ref<QRState>({
  open: false,
  source: '',
  label: '',
  image: '',
  status: '',
  loading: false,
  error: '',
})
let qrKey = ''
let qrTimer: ReturnType<typeof setInterval> | null = null
let qrPolling = false

// 计算某平台可用的扫码登录按钮（QQ 额外支持微信扫码）。
function qrButtonsFor(id: string): { source: string; label: string }[] {
  const btns: { source: string; label: string }[] = []
  if (qrSources.value.includes(id)) btns.push({ source: id, label: '扫码登录' })
  if (id === 'qq' && qrSources.value.includes('qq_wx')) {
    btns.push({ source: 'qq_wx', label: '微信扫码' })
  }
  return btns
}

function stopQRPoll() {
  if (qrTimer) {
    clearInterval(qrTimer)
    qrTimer = null
  }
  qrPolling = false
}

async function startQR(source: string, label: string, platformName: string) {
  stopQRPoll()
  qr.value = {
    open: true,
    source,
    label: `${platformName} · ${label}`,
    image: '',
    status: source === 'qq_wx' ? '正在生成二维码，请使用微信扫码' : '正在生成二维码…',
    loading: true,
    error: '',
  }
  qrKey = ''
  try {
    const session = await CreateQRLogin(source)
    qrKey = String(session.key || '')
    qr.value.image = String(session.image_url || '')
    qr.value.loading = false
    if (!qr.value.image) {
      qr.value.error = '二维码生成失败，请重试'
      return
    }
    qr.value.status = source === 'qq_wx' ? '请打开微信扫码登录' : '请打开对应音乐 App 扫码登录'
    qrTimer = setInterval(pollQR, 2200)
  } catch (e) {
    qr.value.loading = false
    qr.value.error = e instanceof Error ? e.message : '二维码创建失败'
  }
}

async function pollQR() {
  if (qrPolling || !qrKey || !qr.value.open) return
  qrPolling = true
  try {
    const result = await CheckQRLogin(qr.value.source, qrKey)
    const status = String(result.status || '')
    if (status === 'success') {
      const target = (result.extra && result.extra.cookie_source) || qrLoginCookieTarget(qr.value.source)
      const cookie = String(result.cookie || '')
      if (cookie) {
        cookies.value = { ...cookies.value, [target]: cookie }
        await persistCookies()
      }
      qr.value.status = '登录成功，Cookie 已自动保存'
      stopQRPoll()
      setTimeout(() => {
        qr.value.open = false
      }, 900)
    } else if (status === 'scanned') {
      qr.value.status = '已扫码，请在手机上确认登录'
    } else if (status === 'expired') {
      qr.value.status = '二维码已过期，请刷新'
      qr.value.error = '二维码已过期'
      stopQRPoll()
    } else if (status === 'failed') {
      qr.value.status = result.message || '登录失败'
      qr.value.error = result.message || '登录失败'
      stopQRPoll()
    } else {
      qr.value.status = qr.value.source === 'qq_wx' ? '等待微信扫码…' : '等待扫码…'
    }
  } catch {
    // 单次轮询失败忽略，等待下一次
  } finally {
    qrPolling = false
  }
}

function qrLoginCookieTarget(source: string): string {
  return source === 'qq_wx' ? 'qq' : source
}

function refreshQR() {
  const [platformName, label] = qr.value.label.split(' · ')
  startQR(qr.value.source, label || '扫码登录', platformName || '')
}

function closeQR() {
  stopQRPoll()
  qr.value.open = false
}
</script>

<template>
  <div class="online-settings">
    <header class="header">
      <div class="title-row">
        <h1>在线设置</h1>
        <button class="close-btn" title="返回" @click="emit('close')">✕</button>
      </div>
      <p class="subtitle">配置音乐平台账户与播放换源策略</p>
    </header>

    <div class="body">
      <section class="card">
        <div class="card-head">
          <h2>音乐账户（Cookie 登录）</h2>
          <span class="hint">粘贴对应平台的登录 Cookie 或直接扫码登录，即可获得更完整的搜索 / 高音质 / 个人歌单能力，仅本地存储。修改后自动保存。</span>
        </div>

        <div v-if="platforms.length === 0" class="empty">加载平台列表失败</div>

        <div v-for="p in platforms" :key="p.id" class="platform-row">
          <div class="platform-meta">
            <span class="platform-name">{{ p.name }}</span>
            <span :class="['badge', statusOf(p.id)]">
              {{ statusOf(p.id) === 'set' ? '已登录' : '未登录' }}
            </span>
            <span class="spacer"></span>
            <button
              v-for="btn in qrButtonsFor(p.id)"
              :key="btn.source"
              class="qr-btn"
              @click="startQR(btn.source, btn.label, p.name)"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
                <rect x="3" y="3" width="7" height="7" rx="1" />
                <rect x="14" y="3" width="7" height="7" rx="1" />
                <rect x="3" y="14" width="7" height="7" rx="1" />
                <path d="M14 14h3v3M20 14v.01M14 20v.01M17 20h.01M20 17h.01M20 20h.01" />
              </svg>
              {{ btn.label }}
            </button>
          </div>
          <textarea
            v-model="cookies[p.id]"
            class="cookie-input"
            rows="2"
            :placeholder="`粘贴 ${p.name} 的 Cookie（形如 key1=val1; key2=val2）`"
            @input="onCookieInput"
          ></textarea>
        </div>

        <div class="save-hint">
          <span :class="['auto-saved', { show: savedFlag }]">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="20 6 9 17 4 12" />
            </svg>
            已自动保存
          </span>
        </div>
      </section>

      <section class="card">
        <div class="card-head">
          <h2>播放换源</h2>
        </div>
        <div class="switch-row">
          <div class="switch-text">
            <span class="switch-title">自动选择无效音源并批量换源</span>
            <span class="hint">播放失败时自动切换到其他可用音源；在在线音乐搜索结果中可一键批量换源。</span>
          </div>
          <ToggleSwitch
            :model-value="settings.autoSwitchInvalidSource"
            @update:model-value="onAutoSwitchChange"
          />
        </div>
      </section>
    </div>

    <!-- 扫码登录弹窗 -->
    <div v-if="qr.open" class="qr-mask" @click.self="closeQR">
      <div class="qr-modal">
        <div class="qr-modal-head">
          <span class="qr-title">{{ qr.label }}</span>
          <button class="close-btn" title="关闭" @click="closeQR">✕</button>
        </div>
        <div class="qr-body">
          <div class="qr-canvas">
            <div v-if="qr.loading" class="qr-spinner"></div>
            <img v-else-if="qr.image" :src="qr.image" alt="登录二维码" />
            <div v-else class="qr-fail">二维码生成失败</div>
          </div>
          <p :class="['qr-status', { error: qr.error }]">{{ qr.error || qr.status }}</p>
          <button class="qr-refresh" @click="refreshQR">刷新二维码</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.online-settings {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.header {
  padding: 22px 28px 14px;
  flex-shrink: 0;
}

.title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.title-row h1 {
  margin: 0;
  font-size: 22px;
  font-weight: 700;
}

.close-btn {
  width: 30px;
  height: 30px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--fluent-text-secondary);
  font-size: 15px;
  cursor: pointer;
  transition: background 0.18s ease;
}

.close-btn:hover {
  background: var(--fluent-bg-hover);
  color: var(--fluent-text);
}

.subtitle {
  margin: 6px 0 0;
  font-size: 13px;
  color: var(--fluent-text-secondary);
}

.body {
  flex: 1;
  overflow-y: auto;
  padding: 4px 28px 28px;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.card {
  border: 1px solid var(--fluent-border);
  border-radius: 14px;
  background: var(--fluent-bg-glass);
  padding: 18px 20px;
}

.card-head {
  margin-bottom: 14px;
}

.card-head h2 {
  margin: 0 0 4px;
  font-size: 15px;
  font-weight: 600;
}

.hint {
  font-size: 12px;
  color: var(--fluent-text-secondary);
  line-height: 1.5;
}

.empty {
  font-size: 13px;
  color: var(--fluent-text-secondary);
  padding: 8px 0;
}

.platform-row {
  margin-bottom: 14px;
}

.platform-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 6px;
}

.platform-name {
  font-size: 13px;
  font-weight: 500;
}

.spacer {
  flex: 1;
}

.badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 10px;
}

.badge.set {
  background: rgba(46, 160, 67, 0.18);
  color: #5fd17e;
}

.badge.empty {
  background: var(--fluent-bg-active);
  color: var(--fluent-text-secondary);
}

.qr-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  height: 26px;
  padding: 0 10px;
  border: 1px solid var(--fluent-border);
  border-radius: 13px;
  background: transparent;
  color: var(--fluent-text);
  font-size: 12px;
  cursor: pointer;
  transition: background 0.18s ease, border-color 0.18s ease;
}

.qr-btn:hover {
  background: var(--fluent-bg-hover);
  border-color: var(--fluent-accent);
}

.qr-btn svg {
  width: 14px;
  height: 14px;
}

.cookie-input {
  width: 100%;
  resize: vertical;
  border: 1px solid var(--fluent-input-border);
  border-radius: 10px;
  background: var(--fluent-input-bg);
  color: var(--fluent-text);
  font-size: 12px;
  font-family: system-ui, -apple-system, "Segoe UI", "Microsoft YaHei", "PingFang SC", sans-serif;
  padding: 8px 10px;
  outline: none;
  transition: border-color 0.18s ease, box-shadow 0.18s ease;
}

.cookie-input:focus {
  border-color: var(--fluent-accent);
  box-shadow: 0 0 0 3px rgba(0, 120, 212, 0.2);
}

.save-hint {
  height: 20px;
  display: flex;
  align-items: center;
  margin-top: 2px;
}

.auto-saved {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  color: #5fd17e;
  opacity: 0;
  transform: translateY(2px);
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.auto-saved.show {
  opacity: 1;
  transform: translateY(0);
}

.auto-saved svg {
  width: 14px;
  height: 14px;
}

.switch-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.switch-text {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.switch-title {
  font-size: 13px;
  font-weight: 500;
}

/* 扫码登录弹窗 */
.qr-mask {
  position: fixed;
  inset: 0;
  z-index: 60;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(4px);
}

.qr-modal {
  width: 340px;
  border-radius: 16px;
  background: var(--fluent-bg-glass);
  border: 1px solid var(--fluent-border);
  box-shadow: 0 18px 50px rgba(0, 0, 0, 0.35);
  overflow: hidden;
}

.qr-modal-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px;
  border-bottom: 1px solid var(--fluent-border);
}

.qr-title {
  font-size: 14px;
  font-weight: 600;
}

.qr-body {
  padding: 22px 18px 24px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
}

.qr-canvas {
  width: 200px;
  height: 200px;
  border-radius: 12px;
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.qr-canvas img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.qr-fail {
  color: #888;
  font-size: 13px;
}

.qr-spinner {
  width: 34px;
  height: 34px;
  border: 3px solid rgba(0, 0, 0, 0.12);
  border-top-color: var(--fluent-accent);
  border-radius: 50%;
  animation: qr-spin 0.9s linear infinite;
}

@keyframes qr-spin {
  to {
    transform: rotate(360deg);
  }
}

.qr-status {
  margin: 0;
  font-size: 13px;
  color: var(--fluent-text-secondary);
  text-align: center;
}

.qr-status.error {
  color: #ff6b6b;
}

.qr-refresh {
  height: 32px;
  padding: 0 18px;
  border: 1px solid var(--fluent-border);
  border-radius: 16px;
  background: transparent;
  color: var(--fluent-text);
  font-size: 13px;
  cursor: pointer;
  transition: background 0.18s ease, border-color 0.18s ease;
}

.qr-refresh:hover {
  background: var(--fluent-bg-hover);
  border-color: var(--fluent-accent);
}
</style>
