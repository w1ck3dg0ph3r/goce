<script setup lang="ts"></script>

<template>
  <button class="button">
    <slot />
  </button>
</template>

<style scoped lang="scss">
@use '@/assets/themes/theme.scss';

.button {
  display: flex;
  align-items: center;
  gap: 0.2rem;

  @include theme.font('heading');
  color: theme.$textColor;
  font-size: 0.9rem;

  border: 1px solid rgba(0, 0, 0, 0);
  border-radius: 3px;
  background-color: theme.$buttonColor;
  &:hover {
    background-color: theme.$buttonColorHover;
  }
  &:focus {
    border: 1px solid theme.$buttonColorFocus;
    outline: none;
  }

  cursor: pointer;
  :deep(i) {
    font-size: 1rem;
  }
}
</style>
