<script setup lang="ts">
import { keyBy } from 'lodash'
import { computed, ref } from 'vue'

interface Option {
  value: string
  text?: string
}

const props = defineProps<{
  modelValue: string
  options?: Option[]
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const $dropdown = ref<HTMLElement | null>(null)
const options = computed(() => {
  return keyBy(props.options, (o) => o.value)
})
let menuVisible = ref(false)
const selectedText = computed(() => {
  return options.value[props.modelValue]?.text || ''
})

function toggleMenu() {
  menuVisible.value ? closeMenu() : openMenu()
}

function openMenu() {
  menuVisible.value = true
}

function closeMenu() {
  menuVisible.value = false
}

function selectOption(value: string) {
  emit('update:modelValue', value)
  closeMenu()
}
</script>

<template>
  <div
    ref="$dropdown"
    class="dropdown"
    :class="{ open: menuVisible }"
    @click="toggleMenu"
    @blur="closeMenu"
    tabindex="0"
  >
    <div class="button">
      <div class="value" v-text="selectedText"></div>
      <i class="codicon" :class="menuVisible ? 'codicon-triangle-up' : 'codicon-triangle-down'"></i>
    </div>
    <div class="menu" v-show="menuVisible">
      <div
        class="option"
        :class="{ active: o.value == props.modelValue }"
        v-for="o of options"
        :key="o.value"
        @click.stop="selectOption(o.value)"
      >
        {{ o.text }}
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
@use '@/assets/themes/theme.scss';

@use 'sass:color';

$width: 13rem;
$maxMenuHeight: 24rem;
$fontSize: 0.9rem;
$borderRadius: 3px;

.dropdown {
  position: relative;
  min-width: $width;

  @include theme.font('heading');
  color: theme.$textColor;
  font-size: $fontSize;
  cursor: pointer;

  .button {
    width: 100%;
    height: 100%;
    padding: 0 0.5rem 0 0.5rem;
    display: flex;
    align-items: center;
    gap: 0.5rem;

    border: 1px solid transparent;
    border-radius: $borderRadius;

    background-color: theme.$buttonColor;

    &:hover {
      background-color: theme.$buttonColorHover;
    }

    .value {
      flex: 1;
    }
  }

  &:focus .button {
    border: 1px solid theme.$buttonColorFocus;
    outline: none;
  }

  &.open {
    .button {
      background-color: theme.$buttonColor;
      border-bottom: none;
      border-bottom-left-radius: 0;
      border-bottom-right-radius: 0;
    }
  }

  .menu {
    position: absolute;
    width: $width;
    max-height: $maxMenuHeight;
    overflow-y: auto;
    z-index: 10;
    background-color: theme.$buttonColor;
    border: 1px solid theme.$buttonColorFocus;
    border-top: none;
    border-bottom-left-radius: $borderRadius;
    border-bottom-right-radius: $borderRadius;
    display: flex;
    flex-direction: column;

    .option {
      color: theme.$textColor;
      font-size: $fontSize;
      cursor: pointer;
      padding: 0.25rem 0 0.25rem 0.5rem;
      background-color: theme.$buttonColor;
      &:hover {
        background-color: theme.$buttonColorHover;
      }
      &.active {
        border-top: 1px solid theme.$buttonColorFocus;
        border-bottom: 1px solid theme.$buttonColorFocus;
      }
    }
  }
}
</style>
