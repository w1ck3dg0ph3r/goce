<script setup lang="ts"></script>

<template>
  <button class="menu-button">
    <slot></slot>
  </button>
</template>

<style scoped lang="scss">
@use 'sass:color';

@use '@/assets/themes/theme.scss';

.menu-button {
  display: flex;
  align-items: center;
  gap: 0.2rem;

  border: none;
  border-radius: 3px;
  background-color: theme.$buttonColor;
  opacity: 0.7;
  &:hover {
    opacity: 1;
  }
  &:focus {
    outline: 2px solid theme.$buttonColor;
  }

  color: theme.$textColor;
  font-size: 1rem;
  cursor: pointer;
  :deep(i) {
    font-size: 1rem;
  }
}
</style>
