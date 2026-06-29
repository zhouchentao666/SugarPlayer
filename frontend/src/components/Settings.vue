<script lang="ts" setup>
import { type AppSettings } from '../composables/useConfig'
import SettingCard from './settings/SettingCard.vue'
import SettingRow from './settings/SettingRow.vue'
import SegmentedControl from './settings/SegmentedControl.vue'
import ColorPicker from './settings/ColorPicker.vue'
import ToggleSwitch from './settings/ToggleSwitch.vue'
import WindowEffectSettings from './settings/WindowEffectSettings.vue'

const props = defineProps<{
  settings: AppSettings
}>()

const emit = defineEmits<{
  (e: 'update:settings', settings: AppSettings): void
  (e: 'close'): void
}>()

function update(partial: Partial<AppSettings>) {
  emit('update:settings', { ...props.settings, ...partial })
}

const themes = [
  { value: 'system', label: '跟随系统' },
  { value: 'light', label: '浅色' },
  { value: 'dark', label: '深色' },
] as const

const qualities = [
  { value: 'standard', label: '标准' },
  { value: 'high', label: '高品质' },
  { value: 'lossless', label: '无损' },
] as const

const accentColors = ['#0078d4', '#107c10', '#ff8c00', '#d13438', '#881798', '#00b7c3']

function handleQualityChange(event: Event) {
  const value = (event.target as HTMLSelectElement).value as AppSettings['quality']
  update({ quality: value })
}
</script>

<template>
  <div class="settings">
    <div class="settings-header">
      <h1>设置</h1>
      <button class="close-btn" @click="emit('close')">✕</button>
    </div>
    <div class="settings-content">
      <SettingCard title="外观">
        <SettingRow label="应用主题" description="选择应用使用的颜色模式">
          <SegmentedControl
            :options="themes"
            :model-value="settings.theme"
            @update:model-value="value => update({ theme: value as AppSettings['theme'] })"
          />
        </SettingRow>
        <SettingRow label="强调色" description="选择应用使用的强调色">
          <ColorPicker
            :colors="accentColors"
            :model-value="settings.accentColor"
            @update:model-value="value => update({ accentColor: value })"
          />
        </SettingRow>
      </SettingCard>

      <SettingCard title="播放">
        <SettingRow label="默认音质" description="在线播放时的首选音质">
          <select :value="settings.quality" class="fluent-select" @change="handleQualityChange">
            <option v-for="q in qualities" :key="q.value" :value="q.value">{{ q.label }}</option>
          </select>
        </SettingRow>
        <SettingRow label="自动播放" description="启动应用后继续播放上次歌曲">
          <ToggleSwitch
            :model-value="settings.autoplay"
            @update:model-value="value => update({ autoplay: value })"
          />
        </SettingRow>
      </SettingCard>

      <WindowEffectSettings :settings="settings" @update="update" />

      <SettingCard title="关于">
        <SettingRow label="SugarMusic" description="一个简洁的本地音乐播放器">
          <span class="setting-value">v0.0.1</span>
        </SettingRow>
      </SettingCard>
    </div>
  </div>
</template>

<style scoped>
.settings {
  height: 100%;
  padding: 28px 32px;
  color: var(--fluent-text);
  overflow-y: auto;
  box-sizing: border-box;
}

.settings-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 28px;
}

.settings-header h1 {
  margin: 0;
  font-size: 28px;
  font-weight: 600;
  letter-spacing: -0.2px;
}

.close-btn {
  width: 34px;
  height: 34px;
  border: none;
  border-radius: 8px;
  background: var(--fluent-bg-hover);
  color: inherit;
  font-size: 14px;
  cursor: pointer;
  transition: background 0.18s ease;
}

.close-btn:hover {
  background: var(--fluent-bg-active);
}

.settings-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
  max-width: 720px;
}

.fluent-select {
  padding: 6px 10px;
  border-radius: 6px;
  border: 1px solid var(--fluent-border);
  background: var(--fluent-bg-hover);
  color: var(--fluent-text);
  font-size: 13px;
  outline: none;
  cursor: pointer;
}

.setting-value {
  font-size: 13px;
  color: var(--fluent-text-secondary);
  white-space: nowrap;
}
</style>
