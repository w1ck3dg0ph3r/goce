<script lang="ts">
import { onMounted, type InjectionKey, type Ref, provide, ref, readonly, reactive } from 'vue'

export interface TabData {
  title: string
}

export const TabsInjectionKey = Symbol('tabs') as InjectionKey<{
  addTab: (tabData: TabData) => number
  removeTab: (idx: number) => void
  activeTab: Readonly<Ref<number>>
}>
</script>

<script setup lang="ts">
const activeTab = ref(0)
const tabs = reactive(new Array<TabData>())

provide(TabsInjectionKey, {
  addTab: (tabData) => {
    tabs.push(tabData)
    return tabs.length - 1
  },
  removeTab: (idx) => {
    tabs.splice(idx, 1)
  },
  activeTab: readonly(activeTab),
})

function selectTab(idx: number) {
  activeTab.value = idx
}

onMounted(() => {})
</script>

<template>
  <div class="tabs-container">
    <div class="tabs-header">
      <button class="tab-button"
        v-for="(t, i) of tabs"
        :key="t.title"
        :class="{ active: i == activeTab }"
        @click="selectTab(i)"
      >
        {{ t.title }}
      </button>
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
    button.tab-button {
      border: none;
      padding: 0.5rem;
      cursor: pointer;
      border-top-left-radius: 3px;
      border-top-right-radius: 3px;
      background-color: theme.$backgroundColorHighlight;
      &.active {
        background-color: theme.$buttonColor;
      }
      @include theme.font('heading');
      color: theme.$textColor;
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
