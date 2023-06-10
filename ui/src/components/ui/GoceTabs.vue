<script lang="ts">
import { type InjectionKey, type Ref, provide, ref, readonly, reactive, nextTick } from 'vue'

export interface TabData {
  id: symbol
  title: string
}

export const TabsInjectionKey = Symbol('tabs') as InjectionKey<{
  addTab(tabData: TabData): void
  removeTab(tabId: symbol): void
  activeTab: Readonly<Ref<symbol | null>>
}>
</script>

<script setup lang="ts">
const tabsOrder = reactive(new Array<symbol>())
const tabs = reactive(new Map<symbol, TabData>())
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
}>()

provide(TabsInjectionKey, {
  addTab,
  removeTab,
  activeTab: readonly(activeTabId),
})

function addTab(tabData: TabData) {
  tabs.set(tabData.id, tabData)
  tabsOrder.push(tabData.id)
  activeTabId.value = tabData.id
}

function removeTab(tabId: symbol) {
  const index = tabsOrder.indexOf(tabId)
  if (tabId == activeTabId.value) {
    chooseNextActiveTab(index)
  }
  tabsOrder.splice(index, 1)
  tabs.delete(tabId)
}

function selectTab(id: symbol) {
  activeTabId.value = id
  emit('tabSelected', id)
}

function chooseNextActiveTab(current: number) {
  if (current < tabsOrder.length - 1) {
    activeTabId.value = tabsOrder[current + 1]
  } else if (current > 0) {
    activeTabId.value = tabsOrder[current - 1]
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

function startDragging(id: symbol, ev: MouseEvent) {
  if (renameTabId.value) return

  selectTab(id)
  if (tabsOrder.length == 1) return

  let tabButtons = $tabHeader.value?.children
  if (!tabButtons) return

  let index = tabsOrder.indexOf(id)
  isDraggingStarted.value = true
  dragState.index = index
  dragState.id = id
  dragState.left = (tabButtons[index] as HTMLElement).offsetLeft
  dragState.start = ev.clientX
  document.addEventListener('mousemove', handleDragging)
  document.addEventListener('mouseup', stopDragging)
  handleDragging(ev)
}

function swapTabs(i: number, j: number) {
  ;[tabsOrder[i], tabsOrder[j]] = [tabsOrder[j], tabsOrder[i]]
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
  if (dragState.index < tabsOrder.length - 1) {
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

function isRenaming(tabId: symbol) {
  return props.renameable && tabId == renameTabId.value
}

function getTabContentWidth(id: symbol) {
  let tabButtons = $tabHeader.value?.children
  if (!tabButtons) return
  let tabButton = tabButtons[tabsOrder.indexOf(id)] as HTMLElement
  let width = tabButton.getBoundingClientRect().width
  let style = getComputedStyle(tabButton)
  let padding = parseFloat(style.getPropertyValue('padding-left'))
  return width - padding * 2
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
  let newName = $renameInput.value[0].value || ''
  emit('tabRenamed', renameTabId.value, newName)
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
        v-for="id of tabsOrder"
        :key="id"
        class="tab-button"
        :class="{
          active: id == activeTabId,
          placeholder: isDragging && id == dragState.id,
          renaming: isRenaming(id),
        }"
        @mousedown="startDragging(id, $event)"
        @dblclick="startRenaming(id)"
      >
        <template v-if="isRenaming(id)">
          <input
            ref="$renameInput"
            type="text"
            :value="tabs.get(id)?.title"
            @keydown.enter="finishRenaming"
            @keydown.escape="cancelRenaming"
            @blur="cancelRenaming"
            spellcheck="false"
            autocomplete="off"
          />
        </template>
        <template v-else>
          <span>{{ tabs.get(id)?.title }}</span>
          <i v-if="props.closable" class="codicon codicon-close" @mousedown.stop="closeTab(id)"></i>
        </template>
      </button>

      <button ref="$draggingButton" v-show="isDragging" class="tab-button dragging">
        <span>{{ tabs.get(dragState.id!)?.title }}</span>
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
    flex: 1;
  }
}
</style>
