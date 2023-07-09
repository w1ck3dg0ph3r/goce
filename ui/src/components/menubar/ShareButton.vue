<script setup lang="ts">
import GoceButton from '@/components/ui/GoceButton.vue'
import bus from '@/services/bus'
import State from '@/state'

import { nextTick, ref, watch } from 'vue'

const shareButtonText = ref('Share')
const $sharedLink = ref<HTMLInputElement | null>(null)

const inputWidth = ref(0)

watch(
  () => State.sharedCodeLink,
  (newLink, oldLink) => {
    if (newLink && !oldLink) {
      shareButtonText.value = 'Copy'
      nextTick(() => {
        if ($sharedLink.value) {
          $sharedLink.value.focus()
          $sharedLink.value.select()
          inputWidth.value = $sharedLink.value.scrollWidth
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
  <GoceButton @click="shareCode">
    <i class="codicon codicon-cloud-upload"></i>
    <span v-text="shareButtonText"></span>
    <input
      v-show="State.sharedCodeLink"
      ref="$sharedLink"
      class="shared-link"
      type="text"
      readonly
      spellcheck="false"
      @click.stop
      :value="State.sharedCodeLink"
      :style="{ width: `${inputWidth}px` }"
    />
  </GoceButton>
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
