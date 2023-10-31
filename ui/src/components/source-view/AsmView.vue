<script lang="ts" setup>
import State from '@/state'

import MonacoEditor from '@/components/editor/MonacoEditor.vue'
import * as monaco from 'monaco-editor/esm/vs/editor/editor.api'

import { computed, onMounted, ref } from 'vue'
import type { SourceMap } from '@/components/editor/sourcemap'

const props = defineProps<{
  sourceMap: SourceMap
}>()

const emit = defineEmits<{
  (e: 'lineHovered', lineNumber: number): void
  (e: 'revealSource', assemblyLineNumber: number): void
}>()

defineExpose({
  revealLine,
})

const $editor = ref<InstanceType<typeof MonacoEditor> | null>(null)
let lineHasSource: monaco.editor.IContextKey<boolean>

onMounted(() => {
  let editor = $editor.value!.getEditor()
  lineHasSource = editor.createContextKey('lineHasSource', false)

  editor.addAction({
    id: 'goce-reveal-source',
    label: 'Reveal Source',
    contextMenuGroupId: 'navigation',
    precondition: 'lineHasSource',
    keybindings: [monaco.KeyCode.F4],
    run() {
      let assemblyLine = editor?.getPosition()?.lineNumber
      if (assemblyLine) {
        emit('revealSource', assemblyLine)
        emit('lineHovered', assemblyLine)
      }
    },
  })

  editor.onDidChangeCursorPosition((ev) => {
    lineHasSource.set(props.sourceMap.reverseMap.has(ev.position.lineNumber))
  })
})

function revealLine(lineNumber: number) {
  let editor = $editor.value?.getEditor()
  editor?.revealLine(lineNumber)
  editor?.setPosition({ lineNumber: lineNumber, column: 1 })
  editor?.trigger('unfold', 'editor.unfold', {})
}

function lineHovered(lineNumber: number) {
  emit('lineHovered', lineNumber)
}

const lineNumbersMinChars = computed(() => {
  let min = 0
  for (let a of props.sourceMap.assembly.addresses) {
    if (a.length > min) min = a.length
  }
  return min
})

function lineAddress(lineNumber: number): string {
  return props.sourceMap.assembly.addresses[lineNumber - 1] || ''
}
</script>

<template>
  <MonacoEditor
    ref="$editor"
    :code="props.sourceMap.assembly.code"
    :theme="State.theme"
    language="plan9asm"
    :options="{
      fontSize: 10,
      readOnly: true,
      lineNumbers: lineAddress,
      lineNumbersMinChars: lineNumbersMinChars,
    }"
    @hover="lineHovered"
    :decorations="props.sourceMap.assemblyDecorations"
    :highlights="props.sourceMap.highlightedAssembly"
  ></MonacoEditor>
</template>
