<script lang="ts">
import { computed } from 'vue'
import { type InjectionKey, type Ref, provide, ref, readonly, reactive, nextTick } from 'vue'

export interface TabData {
  id: symbol
  title: string
  icon?: string
  order: number
}

export const TabsInjectionKey = Symbol('tabs') as InjectionKey<{
  addTab(tabData: TabData): void
  removeTab(id: symbol): void
  setTabOrder(id: symbol, order: number): void
  activeTabId: Readonly<Ref<symbol | null>>
}>
</script>

<script setup lang="ts">
const tabs = reactive(new Map<Symbol, TabData>())
let activeTabId: Ref<symbol | null> = ref(null)

let $tabHeader: Ref<HTMLElement | null> = ref(null)

const props = withDefaults(
  defineProps<{
    closable?: boolean
    renameable?: boolean
    newTabButton?: boolean
  }>(),
  {
    closable: false,
    renameable: false,
    newTabButton: false,
  }
)

const emit = defineEmits<{
  (e: 'tabSelected', id: symbol): void
  (e: 'newTabClicked'): void
  (e: 'closeTabClicked', id: symbol): void
  (e: 'tabRenamed', id: symbol, name: string): void
  (e: 'tabsSwapped', a: number, b: number): void
}>()

defineExpose({
  selectTab,
})

provide(TabsInjectionKey, {
  addTab,
  removeTab,
  setTabOrder,
  activeTabId: readonly(activeTabId),
})

const tabList = computed(() => {
  let entries = Array.from(tabs.values())
  entries.sort((a, b) => a.order - b.order)
  return entries
})

function addTab(tab: TabData) {
  tabs.set(tab.id, tab)
  activeTabId.value = tab.id
}

function removeTab(id: symbol) {
  const tab = tabs.get(id)
  if (!tab) return
  if (id == activeTabId.value) {
    chooseNextActiveTab(tab.order)
  }
  tabs.delete(id)
}

function setTabOrder(id: symbol, order: number) {
  const tab = tabs.get(id)
  if (tab) {
    tab.order = order
  }
}

function selectTab(id: symbol) {
  activeTabId.value = id
  emit('tabSelected', id)
}

function chooseNextActiveTab(current: number) {
  if (current < tabList.value.length - 1) {
    activeTabId.value = tabList.value[current + 1].id
  } else if (current > 0) {
    activeTabId.value = tabList.value[current - 1].id
  } else {
    activeTabId.value = null
  }
}

function closeTab(id: symbol) {
  emit('closeTabClicked', id)
}

let $draggingButton: Ref<HTMLElement | null> = ref(null)
let isDraggingStarted = ref(false)
let isDragging = ref(false)
const dragState = reactive<{
  index: number
  id: symbol | null
  left: number
  start: number
}>({
  index: -1,
  id: null,
  left: 0,
  start: 0,
})

const draggingTab = computed(() => {
  return dragState.id ? tabs.get(dragState.id) : undefined
})

function startDragging(id: symbol, ev: MouseEvent) {
  if (renameTabId.value) return

  selectTab(id)
  if (tabs.size == 1) return

  let tabButtons = $tabHeader.value?.children
  if (!tabButtons) return

  const tab = tabs.get(id)
  if (!tab) return
  isDraggingStarted.value = true
  dragState.index = tab.order
  dragState.id = tab.id
  dragState.left = (tabButtons[tab.order] as HTMLElement).offsetLeft
  dragState.start = ev.clientX
  document.addEventListener('mousemove', handleDragging)
  document.addEventListener('mouseup', stopDragging)
  handleDragging(ev)
}

function swapTabs(i: number, j: number) {
  emit('tabsSwapped', i, j)
}

function handleDragging(ev: MouseEvent) {
  let el = $draggingButton.value
  if (!isDraggingStarted.value || !el) return

  let tabButtons = $tabHeader.value?.children
  if (!tabButtons) return

  let delta = ev.clientX - dragState.start
  if (Math.abs(delta) > 2) {
    isDragging.value = true
  }

  let left = dragState.left + delta
  let right = left + el.clientWidth
  if (dragState.index > 0) {
    let prev = tabButtons[dragState.index - 1] as HTMLElement
    let middle = prev.offsetLeft + prev.clientWidth * 0.25
    if (left < middle) {
      swapTabs(dragState.index, dragState.index - 1)
      dragState.index = dragState.index - 1
    }
  }
  if (dragState.index < tabs.size - 1) {
    let next = tabButtons[dragState.index + 1] as HTMLElement
    let middle = next.offsetLeft + next.clientWidth * 0.75
    if (right > middle) {
      swapTabs(dragState.index, dragState.index + 1)
      dragState.index = dragState.index + 1
    }
  }
  el.style.left = `${left}px`
}

function stopDragging() {
  isDragging.value = false
  isDraggingStarted.value = false
  let el = $draggingButton.value
  if (el) {
    el.style.position = ''
    el.style.marginLeft = ''
  }
  document.removeEventListener('mousemove', handleDragging)
  document.removeEventListener('mouseup', stopDragging)
}

let $renameInput: Ref<HTMLInputElement[]> = ref([])
let renameTabId: Ref<symbol | null> = ref(null)

function isRenaming(id: symbol) {
  return props.renameable && id == renameTabId.value
}

function getTabContentWidth(id: symbol) {
  let tabButtons = $tabHeader.value?.children
  if (!tabButtons) return
  const tab = tabs.get(id)
  if (!tab) return

  let tabButton = tabButtons[tab.order] as HTMLElement
  let width = tabButton.getBoundingClientRect().width
  let style = getComputedStyle(tabButton)
  let padding = parseFloat(style.getPropertyValue('padding-left'))
  let flexGap = parseFloat(style.getPropertyValue('gap'))

  let icon = tabButton.children.item(0)
  let iconOffset = 0
  if (icon && icon?.tagName.toLowerCase() == 'i') {
    iconOffset = icon.getBoundingClientRect().width
    iconOffset += flexGap
  }

  return width - padding * 2 - iconOffset
}

function startRenaming(id: symbol) {
  if (!props.renameable) return
  let inputWidth = getTabContentWidth(id)
  renameTabId.value = id
  nextTick(() => {
    if (!$renameInput.value || $renameInput.value.length == 0) return
    let $input = $renameInput.value[0]
    $input.style.width = `${inputWidth}px`
    $input.setSelectionRange(0, $input.value.length)
    $input.focus()
  })
}

function finishRenaming() {
  if (!renameTabId.value) return
  let newTitle = $renameInput.value[0].value || ''
  emit('tabRenamed', renameTabId.value, newTitle)
  tabs.get(renameTabId.value)!.title = newTitle
  renameTabId.value = null
}

function cancelRenaming() {
  renameTabId.value = null
}
</script>

<template>
  <div class="tabs-container">
    <div ref="$tabHeader" class="tabs-header">
      <button
        v-for="tab of tabList"
        :key="tab.id"
        class="tab-button"
        :class="{
          active: tab.id == activeTabId,
          placeholder: isDragging && tab.id == dragState.id,
          renaming: isRenaming(tab.id),
        }"
        @mousedown="startDragging(tab.id, $event)"
        @dblclick="startRenaming(tab.id)"
      >
        <i v-if="tab.icon" class="codicon" :class="`codicon-${tab.icon}`"></i>
        <template v-if="isRenaming(tab.id)">
          <input
            ref="$renameInput"
            type="text"
            :value="tab.title"
            @keydown.enter="finishRenaming"
            @keydown.escape="cancelRenaming"
            @blur="cancelRenaming"
            spellcheck="false"
            autocomplete="off"
          />
        </template>
        <template v-else>
          <span>{{ tab.title }}</span>
          <i
            v-if="props.closable"
            class="codicon codicon-close"
            @mousedown.stop="closeTab(tab.id)"
          ></i>
        </template>
      </button>

      <button ref="$draggingButton" v-show="isDragging" class="tab-button dragging">
        <i v-if="draggingTab?.icon" class="codicon" :class="`codicon-${draggingTab?.icon}`"></i>
        <span>{{ draggingTab?.title }}</span>
        <i v-if="props.closable" class="codicon codicon-close"></i>
      </button>

      <button v-if="props.newTabButton" class="tab-button" @click="emit('newTabClicked')">+</button>

      <slot name="buttons"></slot>
      <div class="spacer"></div>
      <slot name="buttons-right"></slot>
    </div>

    <div class="tabs-content">
      <slot></slot>
    </div>
  </div>
</template>

<style scoped lang="scss">
@use '@/assets/themes/theme.scss';

.tabs-container {
  display: flex;
  flex-direction: column;

  .tabs-header {
    display: flex;
    background-color: theme.$backgroundColor;
    border-bottom: 1px solid theme.$buttonColor;

    button.tab-button {
      display: flex;
      align-items: center;
      gap: 0.2rem;
      border: none;
      border-bottom: none;
      padding: 0.2rem 0.5rem 0.2rem 0.5rem;

      cursor: pointer;
      @include theme.font('heading');
      color: theme.$textColor;

      border-top-left-radius: 5px;
      border-top-right-radius: 5px;
      background-color: theme.$buttonColorInactive;

      &:hover {
        background-color: theme.$buttonColorHover;
      }

      &.active {
        background-color: theme.$buttonColor;
        border-bottom: none;
      }

      &.placeholder {
        background-color: theme.$backgroundColorHighlight;
        & > * {
          opacity: 0;
        }
      }

      &.renaming {
        padding: 0 0.5rem 0 0.5rem;
      }

      &.dragging {
        position: absolute;
        background-color: theme.$buttonColor;
        border-bottom: none;
      }

      input {
        height: 100%;
        background-color: theme.$buttonColor;
        @include theme.font('heading');
        color: theme.$textColor;
        font-size: 0.8rem;
        &::selection {
          background-color: theme.$buttonColorFocus;
        }
      }
    }

    .spacer {
      flex: 1;
    }
  }

  .tabs-content {
    position: relative;
    height: 100%;
  }
}
</style>
