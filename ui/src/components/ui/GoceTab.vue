<script setup lang="ts">
import {
  computed,
  inject,
  onMounted,
  onUnmounted,
  ref,
  type Ref,
  getCurrentInstance,
  onUpdated,
} from 'vue'
import { TabsInjectionKey, type TabData } from './GoceTabs.vue'

const props = defineProps<{
  id?: string
  title: string
}>()

const tabsInjection = inject(TabsInjectionKey)
if (!tabsInjection) throw new Error('tab not inside tabs')

let tabData: Ref<TabData | null> = ref(null)

onMounted(() => {
  tabData.value = {
    id: getCurrentInstance()?.vnode.key as symbol,
    title: props.title,
  }
  tabsInjection.addTab(tabData.value)
})

onUpdated(() => {
  if (tabData.value) {
    tabData.value.title = props.title
  }
})

onUnmounted(() => {
  if (tabData.value?.id) tabsInjection.removeTab(tabData.value.id)
})

const isActive = computed(() => tabsInjection.activeTab.value == tabData.value?.id)
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
