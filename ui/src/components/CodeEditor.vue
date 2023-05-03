<script lang="ts" setup>
import * as monaco from 'monaco-editor/esm/vs/editor/editor.api'
import 'monaco-editor/esm/vs/editor/editor.all'
import 'monaco-editor/esm/vs/basic-languages/go/go.contribution'

import '@/components/editor/monaco-environment'

import { debounce } from 'lodash'

import { computed, onMounted, ref, watch } from 'vue'
import type { SourceMap } from './editor/sourcemap'
import bus from '@/services/bus'
import State from '@/state'

const props = defineProps<{
  defaultCode: string
}>()

const emit = defineEmits(['change'])

defineExpose({
  getCode,
  setCode,
  setMapping,
})

const $editor = ref<HTMLElement | null>(null)
let editor: monaco.editor.IStandaloneCodeEditor
let mapping: SourceMap | null
let unwatchHighligh: () => void | null
let blockDecorations: monaco.editor.IEditorDecorationsCollection
let highlightDecorations: monaco.editor.IEditorDecorationsCollection

const editorTheme = computed(() => {
  if (State.theme == 'dark') return 'vs-dark'
  return 'vs'
})

watch(editorTheme, (newTheme) => {
  editor.updateOptions({
    theme: newTheme,
  })
})

onMounted(() => {
  createEditor()
  layoutEditor()
  editor.focus()
})

function createEditor() {
  editor = monaco.editor.create($editor.value!, {
    value: props.defaultCode,
    theme: editorTheme.value,
    language: 'go',
    insertSpaces: false,
    tabSize: 2,
    // minimap: { enabled: false },
    scrollbar: {
      useShadows: false,
    },
    scrollBeyondLastLine: false,
  })

  blockDecorations = editor.createDecorationsCollection()
  highlightDecorations = editor.createDecorationsCollection()

  editor.onMouseMove((ev) => {
    if (!mapping) return
    if (!ev.target.position) return
    const lineNum = ev.target.position.lineNumber
    mapping.highlightFromSource(lineNum)
  })

  editor.onDidChangeModelContent(
    debounce(() => {
      emit('change', editor.getValue())
    }, 1000)
  )

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

function getCode(): string {
  return editor?.getModel()?.getValue() ?? ''
}

function setCode(code: string, keepCursor: boolean = true) {
  let pos: monaco.Position = new monaco.Position(1, 1)
  if (keepCursor) pos = editor.getPosition() ?? pos
  editor.getModel()?.setValue(code)
  if (keepCursor) editor.setPosition(pos)
}

function setMapping(m: SourceMap) {
  if (unwatchHighligh) unwatchHighligh()
  unwatchHighligh = watch(m.highlightedRange, (v) => {
    if (v < 0) {
      highlightDecorations.clear()
      return
    }
    highlightDecorations.set([
      {
        range: new monaco.Range(v, 1, v, 1),
        options: {
          isWholeLine: true,
          linesDecorationsClassName: 'line-highlight',
          blockClassName: 'block-highlight',
        },
      },
    ])
  })
  mapping = m
  blockDecorations.set(m.sourceBlockDecorations())
  highlightDecorations.clear()
}
</script>

<template>
  <div class="code-editor">
    <div ref="$editor"></div>
  </div>
</template>

<style lang="scss" scoped>
@use 'sass:color';

.code-editor {
  overflow: hidden;
}

:deep(.line-highlight) {
  background-color: #007acc;
  width: 5px !important;
  margin-left: 3px;
}

:deep(.block-highlight) {
  background-color: color.change(#007acc, $alpha: 0.2);
}
</style>
