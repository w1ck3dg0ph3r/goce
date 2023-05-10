<script setup lang="ts">
import MenuButton from './MenuButton.vue'
import bus from '@/services/bus'
import State from '@/state'

import { nextTick, ref, watch } from 'vue'

const shareButtonText = ref('Share')
const $sharedLink = ref<HTMLInputElement | null>(null)

watch(
  () => State.sharedCodeLink,
  (newLink, oldLink) => {
    if (newLink && !oldLink) {
      shareButtonText.value = 'Copy'
      nextTick(() => {
        if ($sharedLink.value) {
          $sharedLink.value.focus()
          $sharedLink.value.select()
        }
      })
    }
  }
)

async function shareCode() {
  if (State.sharedCodeLink) {
    navigator.clipboard.writeText(State.sharedCodeLink)
    shareButtonText.value = 'Done'
    setTimeout(() => {
      State.sharedCodeLink = ''
      shareButtonText.value = 'Share'
    }, 1000)
  } else {
    bus.emit('shareCode')
  }
}
</script>

<template>
  <MenuButton @click="shareCode">
    <i class="codicon codicon-cloud-upload"></i>
    <span v-text="shareButtonText"></span>
    <input
      v-show="State.sharedCodeLink"
      ref="$sharedLink"
      class="shared-link"
      type="text"
      spellcheck="false"
      @click.stop
      :value="State.sharedCodeLink"
    />
  </MenuButton>
</template>

<style scoped lang="scss">
@use '@/assets/themes/theme.scss';

input.shared-link {
  background-color: theme.$buttonColor;
  color: theme.$textColor;
  font-size: 0.7rem;
  &::selection {
    background-color: theme.$buttonColorFocus;
  }
}
</style>
