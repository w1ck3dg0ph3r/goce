<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  value?: string
}>()

const emit = defineEmits<{
  (e: 'jumpToSource', line: number, column?: number): void
}>()

interface Line {
  link?: {
    text: string
    line: number
    column: number
  }
  text: string
}

const lines = computed((): Line[] => {
  if (!props.value) return []
  return props.value.split('\n').map((line) => {
    let matches = line.match(/^\.\/main\.go:(\d+):(\d+)/)
    if (matches) {
      return {
        link: {
          text: matches[0],
          line: parseInt(matches[1]),
          column: parseInt(matches[2]),
        },
        text: line.substring(matches[0].length),
      }
    }
    return {
      link: undefined,
      text: line,
    }
  })
})
</script>

<template>
  <div class="build-output">
    <div class="lines">
      <div v-for="(l, i) of lines" :key="i" class="line">
        <span v-if="l.link">
          <a
            :href="l.link.text"
            @click.prevent="emit('jumpToSource', l.link.line, l.link.column)"
            >{{ l.link.text }}</a
          >
        </span>
        <span>{{ l.text }}</span>
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
