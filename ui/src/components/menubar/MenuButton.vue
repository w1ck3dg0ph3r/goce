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

  color: theme.$textColor;
  font-size: 1rem;
  cursor: pointer;
  :deep(i) {
    font-size: 1rem;
  }
}
</style>
