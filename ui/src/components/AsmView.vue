<script lang="ts" setup>
import State from '@/state'
import bus from '@/services/bus'

import MonacoEditor from '@/components/editor/MonacoEditor.vue'

import { ref, watch } from 'vue'

const $editor = ref<InstanceType<typeof MonacoEditor> | null>(null)

watch(
  () => State.sourceMap.assembly,
  (code) => $editor.value?.setValue(code)
)

function lineHovered(lineNumber: number) {
  bus.emit('assemblyLineHovered', lineNumber)
}
</script>

<template>
  <MonacoEditor
    ref="$editor"
    :theme="State.theme"
    language="plan9asm"
    :options="{
      fontSize: 10,
      readOnly: true,
      lineNumbers: 'off',
      folding: false,
    }"
    @hover="lineHovered"
    :decorations="State.sourceMap.assemblyDecorations"
    :highlights="State.sourceMap.highlightedAssembly"
  ></MonacoEditor>
</template>
