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
  codeLeft?: string
  codeRight?: string
  options?: monaco.editor.IStandaloneDiffEditorConstructionOptions
}>()

defineExpose({
  getEditor,
})

const $editor = ref<HTMLElement | null>(null)
let editor: monaco.editor.IStandaloneDiffEditor
let modelLeft: monaco.editor.ITextModel
let modelRight: monaco.editor.ITextModel

let resobs: ResizeObserver

const unsubscribeHandlers = new Array<WatchStopHandle>()
const debouncedLayoutEditor = debounce(layoutEditor, 150)

onMounted(() => {
  createEditor()
  layoutEditor()

  resobs = new ResizeObserver(debouncedLayoutEditor)
  resobs.observe($editor.value!.parentElement!)

  unsubscribeHandlers.push(
    watchEffect(() => {
      monaco.editor.setTheme(editorTheme.value)
    })
  )

  unsubscribeHandlers.push(
    watchEffect(() => {
      if (props.options) editor.updateOptions(props.options)
    })
  )

  unsubscribeHandlers.push(
    watchEffect(() => {
      modelLeft.setValue(props.codeLeft || '')
      modelRight.setValue(props.codeRight || '')
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
  let options: monaco.editor.IStandaloneDiffEditorConstructionOptions = {
    theme: 'vs-dark',
    scrollbar: {
      useShadows: false,
    },
    minimap: {
      enabled: false,
      showSlider: 'always',
    },
    renderWhitespace: 'none',
  }
  options = merge(options, props.options)
  editor = monaco.editor.createDiffEditor($editor.value!, options)

  modelLeft = monaco.editor.createModel(props.codeLeft || '', 'plan9asm')
  modelRight = monaco.editor.createModel(props.codeRight || '', 'plan9asm')

  editor.setModel({
    original: modelLeft,
    modified: modelRight,
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

function getEditor(): monaco.editor.IStandaloneDiffEditor {
  return editor
}
</script>

<template>
  <div class="diff-editor">
    <div ref="$editor"></div>
  </div>
</template>

<style lang="scss" scoped>
$diagSize: 0.5em;

.diff-editor {
  overflow: hidden;

  :deep(.monaco-editor .diagonal-fill) {
    background-size: $diagSize $diagSize;
  }
}
</style>
