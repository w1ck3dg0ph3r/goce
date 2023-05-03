<script setup lang="ts"></script>

<template>
  <button class="menu-button">
    <slot></slot>
  </button>
</template>

<style scoped lang="scss">
@use 'sass:color';

.menu-button {
  display: flex;
  align-items: center;
  gap: 0.2rem;

  border: none;
  border-radius: 3px;
  $color: #007acc;
  background-color: color.change($color, $alpha: 0.7);
  &:hover {
    background-color: $color;
  }
  &:focus {
    outline: 2px solid color.change($color, $alpha: 0.7);
  }

  color: #eee;
  font-size: 1rem;
  cursor: pointer;
  :deep(i) {
    font-size: 1rem;
  }
}
</style>
