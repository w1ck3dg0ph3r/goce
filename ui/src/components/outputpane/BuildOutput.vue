<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  value?: string
}>()

const emit = defineEmits<{
  (e: 'jumpToSource', line: number, column?: number): void
}>()

interface Token {
  text: string
  line?: number
  column?: number
}

type Line = Token[]

const lines = computed((): Line[] => {
  if (!props.value) return []
  return props.value.split('\n').map((line) => {
    const tokens: Token[] = []
    const matches = line.matchAll(/\.\/main\.go:(\d+):(\d+)/g)
    if (matches) {
      let pos = 0
      for (const match of matches) {
        if ((match.index || 0) > pos) {
          tokens.push({ text: line.substring(pos, match.index) })
        }
        tokens.push({
          text: match[0],
          line: parseInt(match[1]),
          column: parseInt(match[2]),
        })
        pos = (match.index || 0) + match[0].length
      }
      if (pos < line.length) {
        tokens.push({ text: line.substring(pos) })
      }
    } else {
      tokens.push({ text: line })
    }
    return tokens
  })
})
</script>

<template>
  <div class="build-output">
    <div class="lines">
      <div v-for="(l, i) of lines" :key="i" class="line">
        <template v-for="t of l" :key="t">
          <span v-if="t.line">
            <a href="" @click.prevent="emit('jumpToSource', t.line, t.column)" v-text="t.text" />
          </span>
          <span v-else v-text="t.text" />
        </template>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
@use '@/assets/themes/theme.scss';

.build-output {
  height: 100%;
  background-color: theme.$editorBackgroundColor;
  color: theme.$editorTextColor;
  position: relative;

  .lines {
    position: absolute;
    overflow-y: scroll;
    left: 0;
    right: 0;
    top: 0;
    bottom: 0;
    padding: 0.25rem;

    font-family: 'Droid Sans Mono', 'monospace', monospace;
    font-size: 10px;
    line-height: 1.25em;

    a {
      color: theme.$editorTextColor;
      font-weight: bold;
      text-decoration: none;
    }
  }
}
</style>
