<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'

const props = defineProps<{
  modelValue: number
  options: string[]
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: number): void
}>()

const $button = ref<HTMLElement | null>(null)
const $menu = ref<HTMLElement | null>(null)
let menuVisible = ref(false)

const selectedText = computed(() => {
  return props.options[props.modelValue] || ''
})

const sizeObserver = new ResizeObserver((entries) => {
  if (entries.length == 0) return
  const b = entries[0]
  if ($menu.value) $menu.value.style.width = b.borderBoxSize[0]?.inlineSize + 'px'
})

onMounted(() => {
  sizeObserver.observe($button.value!)
})

onUnmounted(() => {
  sizeObserver.disconnect()
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

function selectOption(index: number) {
  emit('update:modelValue', index)
  closeMenu()
}
</script>

<template>
  <div
    class="dropdown"
    :class="{ open: menuVisible }"
    @click="toggleMenu"
    @blur="closeMenu"
    tabindex="0"
  >
    <div ref="$button" class="button">
      <div class="value" v-text="selectedText"></div>
      <i class="codicon" :class="menuVisible ? 'codicon-triangle-up' : 'codicon-triangle-down'"></i>
    </div>
    <div ref="$menu" class="menu" v-show="menuVisible">
      <div
        class="option"
        :class="{ active: i == props.modelValue }"
        v-for="(text, i) of options"
        :key="i"
        @click.stop="selectOption(i)"
      >
        {{ text }}
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
@use '@/assets/themes/theme.scss';

@use 'sass:color';

$minWidth: 5rem;
$maxMenuHeight: 24rem;
$fontSize: 0.9rem;
$borderRadius: 3px;

.dropdown {
  position: relative;
  min-width: $minWidth;

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
      text-wrap: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
    }
  }

  &:focus .button {
    border: 1px solid theme.$buttonColorFocus;
    outline: none;
  }

  &.open {
    .button {
      background-color: theme.$buttonColor;
      border-bottom: 1px solid transparent;
      border-bottom-left-radius: 0;
      border-bottom-right-radius: 0;
    }
  }

  .menu {
    position: absolute;
    min-width: $minWidth;
    max-height: $maxMenuHeight;
    overflow-y: auto;
    z-index: 999;
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
      text-wrap: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
      cursor: pointer;
      padding: 0.25rem 0 0.25rem 0.5rem;
      background-color: theme.$buttonColor;
      &:hover {
        background-color: theme.$buttonColorHover;
      }
      &.active {
        background-color: theme.$buttonColorFocus;
      }
    }
  }
}
</style>
