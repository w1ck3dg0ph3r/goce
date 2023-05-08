<script lang="ts">
import 'monaco-editor/esm/vs/editor/editor.all'
import 'monaco-editor/esm/vs/basic-languages/go/go.contribution'

import '@/components/editor/environment'
import '@/components/editor/plan9asm'
</script>

<script setup lang="ts">
import bus from '@/services/bus'

import * as monaco from 'monaco-editor/esm/vs/editor/editor.api'

import { computed, onMounted, onUnmounted, ref, watchEffect, type WatchStopHandle } from 'vue'
import { debounce, merge } from 'lodash'

const props = defineProps<{
  theme: string
  language: string
  value?: string
  options?: monaco.editor.IStandaloneEditorConstructionOptions
  decorations?: monaco.editor.IModelDeltaDecoration[]
  highlights?: Array<monaco.Range>
}>()

const emit = defineEmits<{
  (event: 'change', code: string): void
  (event: 'hover', line: number): void
}>()

defineExpose({
  getEditor,
  getValue,
  setValue,
})

const $editor = ref<HTMLElement | null>(null)
let editor: monaco.editor.IStandaloneCodeEditor
let decorations: monaco.editor.IEditorDecorationsCollection
let highlightDecorations: monaco.editor.IEditorDecorationsCollection
let hoveredLine = -1

const unsubscribeHandlers = new Array<WatchStopHandle>()

onMounted(() => {
  createEditor()
  layoutEditor()

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
    value: props.value,
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

  editor.onDidChangeModelContent(
    debounce(() => {
      emit('change', editor.getValue())
    }, 1000)
  )

  editor.onMouseMove((ev) => {
    const lineNum = ev.target?.position?.lineNumber || -1
    if (lineNum != hoveredLine) {
      hoveredLine = lineNum
      emit('hover', hoveredLine)
    }
  })

  const debounced = debounce(layoutEditor, 300)
  window.addEventListener('resize', debounced)
  unsubscribeHandlers.push(() => window.removeEventListener('resize', debounced))
  bus.on('editorLayoutRequested', layoutEditor)
  unsubscribeHandlers.push(() => bus.off('editorLayoutRequested', layoutEditor))
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

function getValue(): string {
  return editor?.getModel()?.getValue() ?? ''
}

function setValue(code: string, keepCursor: boolean = true) {
  if (!code) return
  let pos: monaco.Position = new monaco.Position(1, 1)
  if (keepCursor) pos = editor.getPosition() ?? pos
  editor.getModel()?.setValue(code)
  if (keepCursor) editor.setPosition(pos)
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
    margin-left: 3px;
  }

  :deep(.block-highlight) {
    background-color: color.change(#007acc, $alpha: 0.2);
  }
}
</style>
