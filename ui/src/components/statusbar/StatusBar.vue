<script setup lang="ts">
import State, { Status } from '@/state'
import { computed } from 'vue'

const isError = computed(() => {
  return State.status == Status.Idle && State.errorMessage != ''
})

const statusText = computed(() => {
  switch (State.status) {
    case Status.Formatting:
      return 'Formatting...'
    case Status.Compiling:
      return 'Compiling...'
    default:
      if (isError.value) return 'Error'
      return ''
  }
})
</script>

<template>
  <div class="status-bar" :class="{ error: isError }">
    <div class="text">{{ statusText }}</div>
  </div>
</template>

<style scoped lang="scss">
@use '@/assets/themes/theme.scss';

$height: 1.57rem;

.status-bar {
  height: $height;
  background-color: theme.$statusBarIdle;
  display: flex;
  gap: 1rem;
  padding-left: 0.5rem;
  &.error {
    background-color: theme.$statusBarError;
  }

  .text {
    line-height: $height;
    color: theme.$statusBarTextColor;
  }
}
</style>
