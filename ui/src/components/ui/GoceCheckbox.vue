<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', v: boolean): void
}>()

const iconName = computed(() => {
  return props.modelValue ? 'codicon-pass-filled' : 'codicon-circle-large'
})

function onClick() {
  emit('update:modelValue', !props.modelValue)
}
</script>

<template>
  <button class="checkbox" @click="onClick">
    <i class="codicon" :class="iconName" />
    <slot></slot>
  </button>
</template>

<style scoped lang="scss">
@use '@/assets/themes/theme.scss';

.checkbox {
  padding: 0;
  display: flex;
  align-items: baseline;
  gap: 0.2rem;

  @include theme.font('heading');
  color: theme.$buttonColor;
  font-size: 0.9rem;

  > i {
    position: relative;
    font-size: 0.9rem;
    top: 0.2rem;
  }

  border: 1px solid rgba(0, 0, 0, 0);
  background-color: theme.$backgroundColor;
  &:focus {
    border-bottom: 1px solid theme.$buttonColorFocus;
    outline: none;
  }

  cursor: pointer;
  :deep(i) {
    font-size: 1rem;
  }
}
</style>
