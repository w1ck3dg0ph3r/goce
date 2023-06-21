<script lang="ts" setup>
import State from '@/state'

import MonacoEditor from '@/components/editor/MonacoEditor.vue'
import * as monaco from 'monaco-editor/esm/vs/editor/editor.api'

import { computed, onMounted, ref } from 'vue'
import type { SourceMap } from './editor/sourcemap'

const cursorPosition = ref(new monaco.Position(1, 1))

const props = defineProps<{
  code?: string
  sourceMap: SourceMap
}>()

const emit = defineEmits<{
  (e: 'update:code', code: string): void
  (e: 'lineHovered', lineNumber: number): void
  (e: 'revealAssembly', sourceLineNumber: number): void
  (e: 'formatCode'): void
}>()

defineExpose({
  revealLine,
  jumpToLocation,
})

const $editor = ref<InstanceType<typeof MonacoEditor> | null>(null)
let lineHasAssembly: monaco.editor.IContextKey<boolean>

onMounted(() => {
  let editor = $editor.value!.getEditor()
  lineHasAssembly = editor.createContextKey('lineHasAssembly', false)

  editor.addAction({
    id: 'goce-reveal-assembly',
    label: 'Reveal Assembly',
    contextMenuGroupId: 'navigation',
    precondition: 'lineHasAssembly',
    keybindings: [monaco.KeyCode.F4],
    run() {
      let sourceLine = editor?.getPosition()?.lineNumber
      if (sourceLine) {
        emit('revealAssembly', sourceLine)
        emit('lineHovered', sourceLine)
      }
    },
  })

  editor.addAction({
    id: 'goce-format-code',
    label: 'Format Code',
    contextMenuGroupId: '1_modification',
    keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyMod.Alt | monaco.KeyCode.KeyF],
    run() {
      emit('formatCode')
    },
  })

  editor.onDidChangeCursorPosition((ev) => {
    lineHasAssembly.set(props.sourceMap.map.has(ev.position.lineNumber))
    cursorPosition.value = ev.position
  })
})

const highlightedLines = computed(() => {
  if (props.sourceMap.highlightedSource) {
    return [props.sourceMap.highlightedSource]
  }
  return undefined
})

function revealLine(line: number) {
  $editor.value?.getEditor().revealLineNearTop(line)
}

function jumpToLocation(line: number, column?: number) {
  $editor.value?.getEditor().focus()
  $editor.value?.getEditor().revealLineInCenter(line)
  $editor.value?.getEditor().setPosition({ lineNumber: line, column: column || 1 })
}

function lineHovered(lineNumber: number) {
  emit('lineHovered', lineNumber)
}
</script>

<template>
  <MonacoEditor
    ref="$editor"
    :code="props.code"
    @update:code="emit('update:code', $event)"
    :theme="State.theme"
    language="go"
    @hover="lineHovered"
    :decorations="props.sourceMap.sourceDecorations"
    :highlights="highlightedLines"
  ></MonacoEditor>
</template>
