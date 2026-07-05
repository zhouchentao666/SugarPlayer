<script lang="ts" setup>
import { onMounted, ref } from 'vue'
import { Version } from '../../bindings/sugarplayer/app'
import { type AppSettings, HOTKEY_ACTIONS, type HotkeyAction } from '../composables/useConfig'
import SettingCard from './settings/SettingCard.vue'
import SettingRow from './settings/SettingRow.vue'
import SegmentedControl from './settings/SegmentedControl.vue'
import ColorPicker from './settings/ColorPicker.vue'
import ToggleSwitch from './settings/ToggleSwitch.vue'
import WindowEffectSettings from './settings/WindowEffectSettings.vue'
import HotkeyInput from './settings/HotkeyInput.vue'

const props = defineProps<{
  settings: AppSettings
}>()

const emit = defineEmits<{
  (e: 'update:settings', settings: AppSettings): void
  (e: 'close'): void
  (e: 'check-update'): void
}>()

const appVersion = ref('')

onMounted(async () => {
  try {
    appVersion.value = await Version()
  } catch {
    appVersion.value = ''
  }
})

function update(partial: Partial<AppSettings>) {
  emit('update:settings', { ...props.settings, ...partial })
}

function updateHotkey(action: HotkeyAction, key: string | undefined) {
  const next = { ...props.settings.hotkeys }
  if (key) {
    next[action] = key
  } else {
    delete next[action]
  }
  update({ hotkeys: next })
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

const fullScreenBackgrounds = [
  { value: 'static', label: '静态' },
  { value: 'dynamic', label: '动态' },
] as const

const coverTransitions = [
  { value: 'fade', label: '淡入淡出' },
  { value: 'slide-left', label: '左边滑入滑出' },
  { value: 'slide-both', label: '左右滑入滑出' },
] as const

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
        <SettingRow label="打开后自动播放音乐" description="启动应用后自动继续播放">
          <ToggleSwitch
            :model-value="settings.autoplay"
            @update:model-value="value => update({ autoplay: value })"
          />
        </SettingRow>
        <SettingRow label="重启后保存播放列表和当前音乐" description="退出时记住当前播放的列表、歌曲和进度">
          <ToggleSwitch
            :model-value="settings.savePlaylistAndSong"
            @update:model-value="value => update({ savePlaylistAndSong: value })"
          />
        </SettingRow>
      </SettingCard>

      <WindowEffectSettings :settings="settings" @update="update" />

      <SettingCard title="全屏播放器">
        <SettingRow label="背景效果" description="选择全屏播放器的背景效果">
          <SegmentedControl
            :options="fullScreenBackgrounds"
            :model-value="settings.fullScreenBackground"
            @update:model-value="value => update({ fullScreenBackground: value as AppSettings['fullScreenBackground'] })"
          />
        </SettingRow>
        <SettingRow label="封面切换动画" description="切换歌曲时封面的过渡效果">
          <SegmentedControl
            :options="coverTransitions"
            :model-value="settings.coverTransition"
            @update:model-value="value => update({ coverTransition: value as AppSettings['coverTransition'] })"
          />
        </SettingRow>
        <SettingRow label="沉浸式播放栏" description="鼠标移开时淡化播放栏中间与右侧，移入时恢复显示">
          <ToggleSwitch
            :model-value="settings.immersivePlayerBar"
            @update:model-value="value => update({ immersivePlayerBar: value })"
          />
        </SettingRow>
      </SettingCard>

      <SettingCard title="窗口">
        <SettingRow label="重启后保存窗口位置和大小" description="退出时记住窗口的位置与尺寸">
          <ToggleSwitch
            :model-value="settings.saveWindowPosition"
            @update:model-value="value => update({ saveWindowPosition: value })"
          />
        </SettingRow>
      </SettingCard>

      <SettingCard title="快捷键">
        <SettingRow
          v-for="action in HOTKEY_ACTIONS"
          :key="action.value"
          :label="action.label"
          description="点击输入框后按下想要的按键"
        >
          <HotkeyInput
            :model-value="settings.hotkeys[action.value]"
            @update:model-value="value => updateHotkey(action.value, value)"
          />
        </SettingRow>
      </SettingCard>

      <SettingCard title="更新">
        <SettingRow label="启动时检查更新" description="每次启动应用时自动检查是否有新版本">
          <ToggleSwitch
            :model-value="settings.checkUpdateOnStartup"
            @update:model-value="value => update({ checkUpdateOnStartup: value })"
          />
        </SettingRow>
        <SettingRow label="检查更新" description="立即检查是否有新版本">
          <button class="fluent-btn" @click="emit('check-update')">立即检查</button>
        </SettingRow>
      </SettingCard>

      <SettingCard title="关于">
        <SettingRow label="SugarMusic" description="一个简洁的本地音乐播放器">
          <span class="setting-value">v{{ appVersion || '0.0.1' }}</span>
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

.fluent-btn {
  padding: 6px 14px;
  border: 1px solid var(--fluent-border);
  border-radius: 6px;
  background: var(--fluent-bg-hover);
  color: var(--fluent-text);
  font-size: 13px;
  cursor: pointer;
  transition: background 0.18s ease;
}

.fluent-btn:hover {
  background: var(--fluent-bg-active);
}
</style>
