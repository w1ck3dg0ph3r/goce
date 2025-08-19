<script setup lang="ts"></script>

<template>
  <div class="loading-indicator">
    <i class="loading-icon codicon codicon-sync" />
  </div>
</template>

<style scoped lang="scss">
@use '@/assets/themes/theme.scss';

.loading-indicator {
  position: absolute;
  width: 100%;
  height: 100%;
  background-color: theme.$backgroundColor;
  opacity: 0.8;

  display: flex;
  align-items: center;
  justify-content: center;

  .loading-icon {
    color: theme.$backgroundColorHighlight;
    font-size: 5rem;
    animation: rotation 2s infinite linear;
  }
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
