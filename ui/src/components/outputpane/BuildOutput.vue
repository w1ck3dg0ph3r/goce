<script lang="ts">
import MonacoEditor from '@/components/editor/MonacoEditor.vue'
import State from '@/state'
import bus from '@/services/bus'

import * as monaco from 'monaco-editor/esm/vs/editor/editor.api'

import { nextTick, ref, watch } from 'vue'

monaco.languages.register({
  id: 'go-build-output',
})

monaco.languages.registerLinkProvider('go-build-output', {
  provideLinks(model) {
    const links = new Array<monaco.languages.ILink>()
    const lines = model.getLinesContent()
    for (let lineNum = 0; lineNum < lines.length; lineNum++) {
      const line = lines[lineNum]
      let matches = line.match(/^\.\/main\.go:(\d+):(\d+)/)
      if (matches) {
        links.push({
          range: new monaco.Range(lineNum + 1, 1, lineNum + 1, matches[0].length + 1),
          url: `jumptosource:${matches[1]}:${matches[2]}`,
        })
      }
    }
    return { links }
  },
})

monaco.editor.registerLinkOpener({
  open(uri) {
    if (uri.scheme == 'jumptosource') {
      let loc = uri.path.split(':')
      bus.emit('jumpToSourceLine', {
        line: parseInt(loc[0]),
        column: parseInt(loc[1]),
      })
      return true
    }
    return false
  },
})
</script>

<script setup lang="ts">
const props = defineProps<{
  value?: string
}>()

const $editor = ref<InstanceType<typeof MonacoEditor> | null>(null)

watch(
  () => props.value,
  (val) => {
    $editor.value?.setValue(val || '')
    nextTick(() => {
      const ed = $editor.value?.getEditor()
      if (ed) {
        ed.setScrollTop(ed.getScrollHeight())
      }
    })
  }
)
</script>

<template>
  <MonacoEditor
    ref="$editor"
    class="output-text"
    :value="props.value"
    language="go-build-output"
    :theme="State.theme"
    :options="{
      fontSize: 12,
      readOnly: true,
      lineNumbers: 'off',
      folding: false,
      minimap: { enabled: false },
      lineDecorationsWidth: 0,
      occurrencesHighlight: false,
      links: true,
    }"
  ></MonacoEditor>
</template>

<style scoped lang="scss">
@use '@/assets/themes/theme.scss';

.output-text {
  width: 100%;
  height: 100%;
}
</style>
