<script lang="ts" setup>
import type { SortMode, ViewMode } from '../composables/usePlaylistView'

const props = defineProps<{
  searchQuery: string
  sortMode: SortMode
  viewMode: ViewMode
  batchMode: boolean
  sortLabels: Record<SortMode, string>
}>()

const emit = defineEmits<{
  'update:searchQuery': [value: string]
  'update:sortMode': [value: SortMode]
  'update:viewMode': [value: ViewMode]
  'toggle-batch': []
}>()

function updateSearch(e: Event) {
  emit('update:searchQuery', (e.target as HTMLInputElement).value)
}

function updateSort(e: Event) {
  emit('update:sortMode', (e.target as HTMLSelectElement).value as SortMode)
}

function setView(mode: ViewMode) {
  emit('update:viewMode', mode)
}
</script>

<template>
  <div class="toolbar">
    <div class="search-box">
      <input
        type="text"
        class="search-input"
        placeholder="搜索歌曲、艺术家、专辑"
        :value="props.searchQuery"
        @input="updateSearch"
      />
    </div>
    <select
      class="control-select"
      :value="props.sortMode"
      @change="updateSort"
    >
      <option
        v-for="(label, mode) in props.sortLabels"
        :key="mode"
        :value="mode"
      >
        {{ label }}
      </option>
    </select>
    <div class="view-toggle">
      <button
        class="view-btn"
        :class="{ active: props.viewMode === 'list' }"
        title="列表视图"
        @click="setView('list')"
      >
        列表
      </button>
      <button
        class="view-btn"
        :class="{ active: props.viewMode === 'grid' }"
        title="网格视图"
        @click="setView('grid')"
      >
        网格
      </button>
    </div>
    <button
      class="batch-btn"
      :class="{ active: props.batchMode }"
      @click="emit('toggle-batch')"
    >
      {{ props.batchMode ? '完成' : '批量' }}
    </button>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
}

.search-box {
  position: relative;
}

.search-input {
  width: 180px;
  padding: 7px 12px;
  border: 1px solid var(--fluent-border);
  border-radius: 8px;
  background: var(--fluent-bg-glass);
  color: var(--fluent-text);
  font-size: 13px;
  outline: none;
  transition: border-color 0.18s ease, background 0.18s ease;
}

.search-input::placeholder {
  color: var(--fluent-text-secondary);
}

.search-input:focus {
  border-color: var(--fluent-accent);
  background: var(--fluent-bg-card);
}

.control-select {
  padding: 7px 10px;
  border: 1px solid var(--fluent-border);
  border-radius: 8px;
  background: var(--fluent-bg-glass);
  color: var(--fluent-text);
  font-size: 13px;
  cursor: pointer;
  outline: none;
}

.view-toggle {
  display: flex;
  border: 1px solid var(--fluent-border);
  border-radius: 8px;
  overflow: hidden;
}

.view-btn {
  padding: 7px 12px;
  border: none;
  background: transparent;
  color: var(--fluent-text-secondary);
  font-size: 13px;
  cursor: pointer;
  transition: background 0.18s ease, color 0.18s ease;
}

.view-btn.active {
  background: var(--fluent-bg-active);
  color: var(--fluent-text);
}

.batch-btn {
  padding: 7px 14px;
  border: 1px solid var(--fluent-border);
  border-radius: 8px;
  background: var(--fluent-bg-glass);
  color: var(--fluent-text);
  font-size: 13px;
  cursor: pointer;
  transition: background 0.18s ease, border-color 0.18s ease;
}

.batch-btn.active {
  border-color: var(--fluent-accent);
  background: var(--fluent-accent);
  color: #fff;
}
</style>
