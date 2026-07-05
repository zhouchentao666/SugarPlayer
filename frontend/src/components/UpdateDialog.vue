<script lang="ts" setup>
import { computed } from 'vue'

export interface UpdateInfo {
  currentVersion: string
  latestVersion: string
  hasUpdate: boolean
  releaseUrl: string
  lanzouUrl: string
  lanzouPassword: string
  error?: boolean
}

const props = defineProps<{
  info: UpdateInfo | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'open', url: string): void
}>()

const visible = computed(() => props.info !== null)

function open(url: string) {
  emit('open', url)
}

function close() {
  emit('close')
}
</script>

<template>
  <Teleport to="body">
    <Transition name="update-fade">
      <div v-if="visible" class="update-overlay" @click.self="close">
        <div class="update-dialog">
          <div class="update-header">
            <span class="update-title">
              {{ info?.error ? '检查更新失败' : (info?.hasUpdate ? '发现新版本' : '已是最新版本') }}
            </span>
            <button class="update-close" @click="close">×</button>
          </div>
          <div class="update-body">
            <template v-if="info?.error">
              <p class="update-version">无法连接到更新服务器</p>
              <p class="update-hint">请检查网络后重试，或稍后再试</p>
            </template>
            <template v-else-if="info?.hasUpdate">
              <p class="update-version">
                当前版本：<strong>{{ info.currentVersion }}</strong>
              </p>
              <p class="update-version">
                最新版本：<strong>{{ info.latestVersion }}</strong>
              </p>
              <p class="update-hint">请选择下载方式：</p>
            </template>
            <template v-else>
              <p class="update-version">
                当前版本：<strong>{{ info?.currentVersion }}</strong>
              </p>
              <p class="update-hint">当前已是最新版本，无需更新</p>
            </template>
          </div>
          <div class="update-actions">
            <template v-if="info?.hasUpdate && !info?.error">
              <button class="update-btn primary" @click="open(info.releaseUrl)">
                GitHub 发布页
              </button>
              <button class="update-btn" @click="open(info.lanzouUrl)">
                蓝奏云（密码 {{ info.lanzouPassword }}）
              </button>
            </template>
            <button class="update-btn secondary" @click="close">
              {{ info?.hasUpdate && !info?.error ? '取消' : '确定' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.update-overlay {
  position: fixed;
  inset: 0;
  z-index: 200;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.35);
  backdrop-filter: blur(4px);
}

.update-dialog {
  width: 360px;
  max-width: 90vw;
  padding: 20px;
  border-radius: 12px;
  background: var(--fluent-bg-card);
  border: 1px solid var(--fluent-border);
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.25);
  color: var(--fluent-text);
}

.update-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.update-title {
  font-size: 18px;
  font-weight: 600;
}

.update-close {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: inherit;
  font-size: 20px;
  cursor: pointer;
  transition: background 0.2s;
}

.update-close:hover {
  background: var(--fluent-bg-hover);
}

.update-body {
  margin-bottom: 20px;
}

.update-version {
  margin: 6px 0;
  font-size: 14px;
  opacity: 0.9;
}

.update-hint {
  margin-top: 14px;
  font-size: 13px;
  opacity: 0.7;
}

.update-actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.update-btn {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid var(--fluent-border);
  border-radius: 8px;
  background: var(--fluent-bg-hover);
  color: var(--fluent-text);
  font-size: 14px;
  cursor: pointer;
  transition: background 0.2s, transform 0.1s;
}

.update-btn:hover {
  background: var(--fluent-bg-active);
}

.update-btn:active {
  transform: scale(0.98);
}

.update-btn.primary {
  background: var(--fluent-accent, #0078d4);
  color: #fff;
  border-color: transparent;
}

.update-btn.primary:hover {
  filter: brightness(1.05);
}

.update-btn.secondary {
  background: transparent;
  opacity: 0.8;
}

.update-fade-enter-active,
.update-fade-leave-active {
  transition: opacity 0.25s ease;
}

.update-fade-enter-from,
.update-fade-leave-to {
  opacity: 0;
}
</style>
