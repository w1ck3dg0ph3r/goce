<script setup lang="ts">
import bus from '@/services/bus'
import '@/components/editor/environment'
import '@/components/editor/plan9asm'

import * as monaco from 'monaco-editor/esm/vs/editor/editor.api'
import 'monaco-editor/esm/vs/editor/editor.all'
import 'monaco-editor/esm/vs/basic-languages/go/go.contribution'

import { computed, onMounted, onUnmounted, ref, watch, watchEffect } from 'vue'
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
  (ev: 'change', code: string): void
  (ev: 'hover', line: number): void
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

onMounted(() => {
  createEditor()
  layoutEditor()
})

onUnmounted(() => {
  editor.dispose()
})

const editorTheme = computed(() => {
  switch (props.theme) {
    case 'light':
      return 'vs'
    case 'dark':
      return 'vs-dark'
    default:
      return 'vs'
  }
})

watch(editorTheme, (newTheme) => {
  editor.updateOptions({
    theme: newTheme,
  })
})

watch(
  () => props.decorations,
  (newDecorations) => {
    if (!newDecorations) {
      decorations.clear()
      return
    }
    decorations.set(newDecorations)
  }
)

watch(
  () => props.highlights,
  (ranges) => {
    if (!ranges) {
      highlightDecorations.clear()
      return
    }
    highlightDecorations.set(
      ranges.map((range) => ({
        range: range,
        options: {
          isWholeLine: true,
          linesDecorationsClassName: 'line-highlight',
          blockClassName: 'block-highlight',
        },
      }))
    )
  }
)

watch(
  () => props.options,
  (newOptions) => {
    if (newOptions) {
      editor.updateOptions(newOptions)
    }
  }
)

function createEditor() {
  let options: monaco.editor.IStandaloneEditorConstructionOptions = {
    value: props.value,
    theme: editorTheme.value,
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
  bus.on('editorLayoutRequested', layoutEditor)
}

function layoutEditor() {
  editor.layout({ width: 0, height: 0 })
  setTimeout(() => {
    const rect = $editor.value!.parentElement!.getBoundingClientRect()
    editor.layout({ width: rect.width, height: rect.height })
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
  <div class="monaco-editor">
    <div ref="$editor"></div>
  </div>
</template>

<style lang="scss" scoped>
@use 'sass:color';

.monaco-editor {
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
