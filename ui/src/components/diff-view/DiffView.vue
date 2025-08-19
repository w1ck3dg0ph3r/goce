<script setup lang="ts">
import State from '@/state'

import type { DiffSettings, SourceTab } from '@/tab'
import DiffPanel from './DiffPanel.vue'

import MonacoDiff from '@/components/editor/MonacoDiff.vue'
import { computed } from 'vue'

const props = defineProps<{
  tabs: Map<symbol, SourceTab>
  settings: DiffSettings
}>()

const emit = defineEmits<{
  (e: 'update:settings', settings: DiffSettings): void
}>()

const codeLeft = computed(() => {
  if (!props.settings.original) return ''
  return props.tabs.get(props.settings.original)?.sourceMap?.assembly?.code || ''
})

const codeRight = computed(() => {
  if (!props.settings.modified) return ''
  return props.tabs.get(props.settings?.modified)?.sourceMap?.assembly?.code || ''
})

function applySettings(settings: DiffSettings) {
  emit('update:settings', settings)
}
</script>

<template>
  <div class="diff-view">
    <DiffPanel :tabs="props.tabs" :settings="settings" @update:settings="applySettings" />
    <MonacoDiff
      ref="$editor"
      class="diff-editor"
      :code-left="codeLeft"
      :code-right="codeRight"
      :theme="State.theme"
      :options="{
        fontSize: 12,
        readOnly: true,
        lineNumbers: 'off',
        automaticLayout: true,
        enableSplitViewResizing: false,
        renderSideBySide: !settings?.inline,
      }"
    />
  </div>
</template>

<style scoped lang="scss">
@use '@/assets/themes/theme.scss';

.diff-view {
  display: flex;
  flex-direction: column;
  background-color: theme.$backgroundColor;
  color-scheme: theme.$colorScheme;

  .diff-editor {
    flex: 1;
  }
}
</style>
