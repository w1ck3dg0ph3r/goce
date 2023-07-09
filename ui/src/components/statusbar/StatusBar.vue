<script lang="ts">
import { computed } from 'vue'

export enum Status {
  Idle,
  Formatting,
  Compiling,
  Error,
}

export interface StatusBarState {
  status: Status
  bottomPaneVisible: boolean
  cursorPosition: {
    lineNumber: number
    column: number
  }
}
</script>

<script setup lang="ts">
const props = defineProps<{
  state: StatusBarState
}>()

const emit = defineEmits<{
  (e: 'toggleBottomPanel'): void
}>()

const isError = computed(() => {
  return props.state.status == Status.Error
})

const statusIcon = computed(() => {
  switch (props.state.status) {
    case (Status.Formatting, Status.Compiling):
      return 'codicon-sync animated'
    case Status.Error:
      return 'codicon-error'
    default:
      return 'codicon-pass'
  }
})

const statusText = computed(() => {
  switch (props.state.status) {
    case Status.Formatting:
      return 'Formatting...'
    case Status.Compiling:
      return 'Compiling...'
  }
  if (isError.value) return 'Error'
  return 'Ready'
})
</script>

<template>
  <div class="status-bar" :class="{ error: isError }">
    <i class="codicon" :class="statusIcon"></i>
    <div class="text">{{ statusText }}</div>
    <div class="spacer"></div>
    <div class="cursor-position">
      Ln {{ props.state.cursorPosition.lineNumber }}, Col {{ props.state.cursorPosition.column }}
    </div>
    <div>
      <button @click="emit('toggleBottomPanel')">
        <i
          class="codicon"
          :class="`codicon-chevron-${props.state.bottomPaneVisible ? 'down' : 'up'}`"
        ></i>
      </button>
    </div>
  </div>
</template>

<style scoped lang="scss">
@use '@/assets/themes/theme.scss';

$height: 1.57rem;

.status-bar {
  height: $height;
  background-color: theme.$statusBarIdle;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding-left: 0.5rem;
  overflow: hidden;

  line-height: $height;
  font-size: 0.75rem;
  color: theme.$statusBarTextColor;
  cursor: default;

  &.error {
    background-color: theme.$statusBarError;
  }

  i {
    line-height: $height;
    font-size: 1rem;
    color: theme.$statusBarTextColor;
    &.animated {
      animation: rotation 2s infinite linear;
    }
  }

  button {
    border: none;
    cursor: pointer;
    background-color: transparent;
    &:hover {
      background-color: rgba($color: #ffffff, $alpha: 0.15);
    }
  }
}

.spacer {
  flex: 1;
}

@keyframes rotation {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(359deg);
  }
}
</style>
