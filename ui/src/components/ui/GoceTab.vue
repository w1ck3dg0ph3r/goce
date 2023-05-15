<script setup lang="ts">
import { computed, inject, onUnmounted } from 'vue'
import { TabsInjectionKey, type TabData } from './GoceTabs.vue'

const props = defineProps<{
  title: string
}>()

const tabsInjection = inject(TabsInjectionKey)
if (!tabsInjection) throw new Error('tab not inside tabs')

const tabData: TabData = { title: props.title }
const tabIdx = tabsInjection.addTab(tabData)

onUnmounted(() => {
  tabsInjection.removeTab(tabIdx)
})

const isActive = computed(() => tabsInjection.activeTab.value == tabIdx)
</script>

<template>
  <div v-show="isActive">
    <slot></slot>
  </div>
</template>
