<script lang="ts" setup>
import State from '@/state'
import bus from '@/services/bus'

import MonacoEditor from '@/components/editor/MonacoEditor.vue'
import * as monaco from 'monaco-editor/esm/vs/editor/editor.api'

import { onMounted, ref, watch } from 'vue'

const $editor = ref<InstanceType<typeof MonacoEditor> | null>(null)
let lineHasSource: monaco.editor.IContextKey<boolean>

onMounted(() => {
  let editor = $editor.value!.getEditor()
  lineHasSource = editor.createContextKey('lineHasSource', false)

  editor.addAction({
    id: 'goce-jump-to-source',
    label: 'Jump To Source',
    contextMenuGroupId: 'navigation',
    precondition: 'lineHasSource',
    keybindings: [monaco.KeyCode.F4],
    run() {
      let assemblyLine = editor?.getPosition()?.lineNumber
      if (assemblyLine) {
        const sourceLine = State.sourceMap.reverseMap.get(assemblyLine)
        if (sourceLine) {
          bus.emit('revealSourceLine', sourceLine)
          bus.emit('sourceLineHovered', sourceLine)
        }
      }
    },
  })

  editor.onDidChangeCursorPosition((ev) => {
    lineHasSource.set(State.sourceMap.reverseMap.has(ev.position.lineNumber))
  })
})

watch(
  () => State.sourceMap.assembly,
  (code) => $editor.value?.setValue(code)
)

bus.on('revealAssemblyLine', (line) => {
  let editor = $editor.value?.getEditor()
  editor?.revealLine(line)
  editor?.setPosition({ lineNumber: line, column: 1 })
  editor?.trigger('unfold', 'editor.unfold', {})
})

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
      // folding: false,
    }"
    @hover="lineHovered"
    :decorations="State.sourceMap.assemblyDecorations"
    :highlights="State.sourceMap.highlightedAssembly"
  ></MonacoEditor>
</template>
