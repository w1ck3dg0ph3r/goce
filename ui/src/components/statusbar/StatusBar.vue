<script setup lang="ts">
import State, { Status } from '@/state'
import { computed } from 'vue'

const isError = computed(() => {
  return State.status == Status.Error
})

const statusIcon = computed(() => {
  switch (State.status) {
    case (Status.Formatting, Status.Compiling):
      return 'codicon-sync animated'
    case Status.Error:
      return 'codicon-error'
    default:
      return 'codicon-pass'
  }
})

const statusText = computed(() => {
  switch (State.status) {
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
      Ln {{ State.cursorPosition.lineNumber }}, Col {{ State.cursorPosition.column }}
    </div>
    <div>
      <button @click="State.bottomPanelVisible = !State.bottomPanelVisible">
        <i
          class="codicon"
          :class="`codicon-chevron-${State.bottomPanelVisible ? 'down' : 'up'}`"
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
