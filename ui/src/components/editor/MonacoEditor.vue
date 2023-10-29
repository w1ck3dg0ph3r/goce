<script lang="ts">
import * as monaco from 'monaco-editor/esm/vs/editor/editor.api'
import 'monaco-editor/esm/vs/editor/editor.all'
import 'monaco-editor/esm/vs/basic-languages/go/go.contribution'

import '@/components/editor/environment'
import '@/components/editor/plan9asm'

import { computed, onMounted, onUnmounted, ref, watchEffect, type WatchStopHandle } from 'vue'
import { debounce, merge } from 'lodash'
</script>

<script setup lang="ts">
const props = defineProps<{
  theme: string
  language: string
  code?: string
  options?: monaco.editor.IStandaloneEditorConstructionOptions
  decorations?: monaco.editor.IModelDeltaDecoration[]
  highlights?: Array<monaco.Range>
}>()

const emit = defineEmits<{
  (event: 'change', code: string): void
  (event: 'update:code', code: string): void
  (event: 'hover', line: number): void
}>()

defineExpose({
  getEditor,
})

const $editor = ref<HTMLElement | null>(null)
let editor: monaco.editor.IStandaloneCodeEditor
let decorations: monaco.editor.IEditorDecorationsCollection
let highlightDecorations: monaco.editor.IEditorDecorationsCollection
let hoveredLine = -1

let resobs: ResizeObserver

const unsubscribeHandlers = new Array<WatchStopHandle>()
const debouncedLayoutEditor = debounce(layoutEditor, 150)

onMounted(() => {
  createEditor()
  layoutEditor()

  editor.onDidChangeModelContent(() => {
    if (props.code != editor.getValue()) {
      emit('update:code', editor.getValue())
    }
  })

  unsubscribeHandlers.push(
    watchEffect(() => {
      if (props.code != editor.getValue()) {
        editor.getModel()?.setValue(props.code ?? '')
        decorations.set(props.decorations || [])
      }
    })
  )

  resobs = new ResizeObserver(debouncedLayoutEditor)
  resobs.observe($editor.value!.parentElement!)

  unsubscribeHandlers.push(
    watchEffect(() => {
      editor.updateOptions({ theme: editorTheme.value })
    })
  )

  unsubscribeHandlers.push(
    watchEffect(() => {
      decorations.set(props.decorations || [])
    })
  )

  unsubscribeHandlers.push(
    watchEffect(() => {
      highlightDecorations.set(
        (props.highlights || []).map((range) => ({
          range: range,
          options: {
            isWholeLine: true,
            linesDecorationsClassName: 'line-highlight',
            blockClassName: 'block-highlight',
          },
        }))
      )
    })
  )

  unsubscribeHandlers.push(
    watchEffect(() => {
      if (props.options) editor.updateOptions(props.options)
    })
  )
})

onUnmounted(() => {
  resobs?.disconnect()
  for (let stop of unsubscribeHandlers) stop()
  unsubscribeHandlers.splice(0)
  editor.dispose()
})

const editorTheme = computed((): string => {
  switch (props.theme) {
    case 'light':
      return 'vs'
    case 'dark':
      return 'vs-dark'
    default:
      return 'vs'
  }
})

function createEditor() {
  let options: monaco.editor.IStandaloneEditorConstructionOptions = {
    theme: 'vs-dark',
    language: props.language,
    insertSpaces: false,
    tabSize: 2,
    scrollbar: {
      useShadows: false,
    },
    scrollBeyondLastLine: false,
  }
  options = merge(options, props.options)
  editor = monaco.editor.create($editor.value!, options)

  decorations = editor.createDecorationsCollection()
  highlightDecorations = editor.createDecorationsCollection()

  editor.onMouseMove((ev) => {
    const lineNum = ev.target?.position?.lineNumber || -1
    if (lineNum != hoveredLine) {
      hoveredLine = lineNum
      emit('hover', hoveredLine)
    }
  })
}

function layoutEditor() {
  editor.layout({ width: 0, height: 0 })
  setTimeout(() => {
    const rect = $editor.value?.parentElement?.getBoundingClientRect()
    if (rect) {
      editor.layout({ width: rect.width, height: rect.height })
    }
  }, 0)
}

function getEditor(): monaco.editor.IStandaloneCodeEditor {
  return editor
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

  :deep(.line-highlight) {
    background-color: #007acc;
    width: 5px !important;
    margin-left: 2px !important;
  }

  :deep(.block-highlight) {
    background-color: color.change(#007acc, $alpha: 0.2);
  }
}
</style>
