<script lang="ts" setup>
import State from '@/state'
import { ref, computed, watch } from 'vue'

import bus from '@/services/bus'

const $textArea = ref<HTMLElement | null>(null)

const visible = computed(() => State.errorMessage != '')
watch(visible, () => {
  bus.emit('editorLayoutRequested')
  setTimeout(() => {
    if ($textArea.value) {
      $textArea.value.scrollTop = $textArea.value.scrollHeight
    }
  })
})
</script>

<template>
  <div id="output-pane" :class="{ hidden: !visible }">
    <textarea ref="$textArea" :value="State.errorMessage" readonly></textarea>
  </div>
</template>

<style lang="scss" scoped>
@use '@/assets/themes/theme.scss';

#output-pane {
  height: 10rem;
  &.hidden {
    display: none;
  }
  padding-top: 0.5rem;
  > textarea {
    background-color: theme.$editorBackgroundColor;
    color: theme.$editorTextColor;
    outline: none;
    border: none;
    resize: none;
    width: 100%;
    height: 100%;
    padding: 0.25rem;
  }
}
</style>
