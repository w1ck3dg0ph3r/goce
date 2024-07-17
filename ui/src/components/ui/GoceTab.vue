<script setup lang="ts">
import { computed, inject, onMounted, onUnmounted } from 'vue'
import { TabsInjectionKey } from './GoceTabs.vue'
import { watch } from 'vue'

const props = defineProps<{
  tabId: symbol
  title: string
  icon?: string
  order: number
}>()

const tabsInjection = inject(TabsInjectionKey)
if (!tabsInjection) throw new Error('tab not inside tabs')

onMounted(() => {
  tabsInjection.addTab({
    id: props.tabId,
    title: props.title,
    icon: props.icon,
    order: props.order,
  })
  watch(
    () => props.order,
    (newOrder) => {
      tabsInjection.setTabOrder(props.tabId, newOrder)
    }
  )
})

onUnmounted(() => {
  tabsInjection.removeTab(props.tabId)
})

const isActive = computed(() => tabsInjection.activeTabId.value == props.tabId)
</script>

<template>
  <div :class="{ hidden: !isActive }" class="tab-content">
    <slot></slot>
  </div>
</template>

<style scoped lang="scss">
.tab-content {
  position: absolute;
  width: 100%;
  height: 100%;
}

.hidden {
  opacity: 0;
  z-index: -9999;
}
</style>
