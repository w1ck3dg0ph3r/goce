<script lang="ts" setup>
import State from '@/state'
import bus from '@/services/bus'

import MonacoEditor from '@/components/editor/MonacoEditor.vue'

import { computed, ref } from 'vue'

defineExpose({
  getCode,
  setCode,
})

const props = defineProps<{
  defaultCode?: string
}>()

const $editor = ref<InstanceType<typeof MonacoEditor> | null>(null)

const highlightedLines = computed(() => {
  if (State.sourceMap.highlightedSource) {
    return [State.sourceMap.highlightedSource]
  }
  return undefined
})

function getCode() {
  return $editor.value?.getValue()
}

function setCode(code: string, keepCursor: boolean) {
  $editor.value?.setValue(code, keepCursor)
}

function lineHovered(lineNumber: number) {
  bus.emit('sourceLineHovered', lineNumber)
}
</script>

<template>
  <MonacoEditor
    ref="$editor"
    :value="props.defaultCode"
    :theme="State.theme"
    language="go"
    @hover="lineHovered"
    :decorations="State.sourceMap.sourceDecorations"
    :highlights="highlightedLines"
  ></MonacoEditor>
</template>
