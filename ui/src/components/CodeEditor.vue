<script lang="ts" setup>
import State from '@/state'
import bus from '@/services/bus'

import MonacoEditor from '@/components/editor/MonacoEditor.vue'
import * as monaco from 'monaco-editor/esm/vs/editor/editor.api'

import { computed, onMounted, ref } from 'vue'

defineExpose({
  getCode,
  setCode,
})

const props = defineProps<{
  defaultCode?: string
}>()

const $editor = ref<InstanceType<typeof MonacoEditor> | null>(null)
let lineHasAssembly: monaco.editor.IContextKey<boolean>

onMounted(() => {
  let editor = $editor.value!.getEditor()
  lineHasAssembly = editor.createContextKey('lineHasAssembly', false)

  editor.addAction({
    id: 'goce-jump-to-assembly',
    label: 'Jump To Assembly',
    contextMenuGroupId: 'navigation',
    precondition: 'lineHasAssembly',
    keybindings: [monaco.KeyCode.F4],
    run() {
      let sourceLine = editor?.getPosition()?.lineNumber
      if (sourceLine) {
        const asmRanges = State.sourceMap.map.get(sourceLine)?.ranges
        if (asmRanges && asmRanges.length > 0) {
          const firstAsmLine = asmRanges[0].start
          bus.emit('jumpToAssemblyLine', firstAsmLine)
          bus.emit('sourceLineHovered', sourceLine)
        }
      }
    },
  })

  editor.onDidChangeCursorPosition((ev) => {
    lineHasAssembly.set(State.sourceMap.map.has(ev.position.lineNumber))
  })
})

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

bus.on('jumpToSourceLine', (line) => {
  $editor.value?.getEditor().revealLineNearTop(line)
})

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
