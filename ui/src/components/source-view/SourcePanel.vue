<script setup lang="ts">
import DropDown from '@/components/ui/DropDown.vue'
import GoceButton from '@/components/ui/GoceButton.vue'
import type { SourceSettings } from './SourceView.vue'

import State from '@/state'

import { computed, reactive } from 'vue'

const props = defineProps<{
  settings: SourceSettings
}>()

const settings = reactive(props.settings)

const compilerOptions = computed(() => {
  return State.compilers.map((c) => ({
    value: c.name,
    text: c.name,
  }))
})

const emit = defineEmits<{
  (e: 'update:settings', settings: SourceSettings): void
  (e: 'format'): void
  (e: 'compile'): void
}>()

function updateCompiler(compiler: string) {
  settings.compiler = compiler
  emit('update:settings', settings)
}
</script>

<template>
  <div class="source-panel">
    <DropDown
      :modelValue="props.settings.compiler"
      @update:modelValue="updateCompiler"
      :options="compilerOptions"
    ></DropDown>
    <div class="spacer"></div>
    <GoceButton @click="emit('format')">
      <i class="codicon codicon-json"></i>
      <span>Format</span>
    </GoceButton>
    <GoceButton @click="emit('compile')">
      <i class="codicon codicon-play"></i>
      <span>Compile</span>
    </GoceButton>
  </div>
</template>

<style scoped lang="scss">
@use '@/assets/themes/theme.scss';

.source-panel {
  height: 2.5rem;
  padding: 0.5rem;
  display: flex;
  gap: 0.5rem;

  .spacer {
    flex: 1;
  }
}
</style>
