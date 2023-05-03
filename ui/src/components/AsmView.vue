<script lang="ts" setup>
import * as monaco from 'monaco-editor/esm/vs/editor/editor.api'
import 'monaco-editor/esm/vs/editor/editor.all'

import '@/components/editor/plan9asm'
import type { SourceMap } from '@/components/editor/sourcemap'
import bus from '@/services/bus'

import { onMounted, ref, watch } from 'vue'
import { debounce } from 'lodash'

const $editor = ref<HTMLElement | null>(null)
let editor: monaco.editor.IStandaloneCodeEditor
let mapping: SourceMap | null
let blockDecorations: monaco.editor.IEditorDecorationsCollection
let highlightDecorations: monaco.editor.IEditorDecorationsCollection

onMounted(() => {
  createEditor()
  layoutEditor()
  editor.focus()
})

defineExpose({
  setAssembly,
  setMapping,
})

function createEditor() {
  editor = monaco.editor.create($editor.value!, {
    value: '',
    language: 'plan9asm',
    fontSize: 10,
    readOnly: true,
    roundedSelection: false,
    lineNumbers: 'off',
    folding: false,
    // minimap: { enabled: false },
    scrollbar: {
      useShadows: false,
    },
    scrollBeyondLastLine: false,
  })

  editor.onMouseMove((ev) => {
    if (!mapping) return
    if (!ev.target.position) return
    const lineNum = ev.target.position.lineNumber
    mapping.highlightFromAssembly(lineNum)
  })

  blockDecorations = editor.createDecorationsCollection()
  highlightDecorations = editor.createDecorationsCollection()

  const debounced = debounce(layoutEditor, 300)
  window.addEventListener('resize', debounced)
  bus.on('editorLayoutRequested', layoutEditor)
}

function layoutEditor() {
  editor.layout({ width: 0, height: 0 })
  setTimeout(() => {
    const rect = $editor.value!.parentElement!.getBoundingClientRect()
    editor.layout({ width: rect.width, height: rect.height })
  }, 0)
}

function setAssembly(asm?: string) {
  editor.getModel()?.setValue(asm ?? '')
}

function setMapping(m: SourceMap) {
  watch(m.highlightedRange, (v) => {
    if (v < 0) {
      highlightDecorations.clear()
      return
    }
    const mapping = m.map.get(v)
    if (!mapping) return
    const decorations = new Array<monaco.editor.IModelDeltaDecoration>()
    for (const range of mapping.ranges) {
      decorations.push({
        range: new monaco.Range(range.start, 1, range.end, 1),
        options: {
          isWholeLine: true,
          linesDecorationsClassName: 'line-highlight',
          blockClassName: 'block-highlight',
        },
      })
    }
    highlightDecorations.set(decorations)
  })
  mapping = m
  blockDecorations.set(m.assemblyBlockDecorations())
}
</script>

<template>
  <div class="asm-view">
    <div ref="$editor"></div>
  </div>
</template>

<style lang="scss" scoped>
@use 'sass:color';

:deep(.line-highlight) {
  background-color: #007acc;
  width: 5px !important;
  margin-left: 3px;
}

:deep(.block-highlight) {
  background-color: color.change(#007acc, $alpha: 0.2);
}
</style>
