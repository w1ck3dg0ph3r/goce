<script lang="ts" setup>
import State from '@/state'
import API from '@/services/api'

import MenuBar from '@/components/menubar/MenuBar.vue'
import StatusBar from '@/components/statusbar/StatusBar.vue'
import GoceTabs from '@/components/ui/GoceTabs.vue'
import GoceTab from '@/components/ui/GoceTab.vue'

import { onMounted, reactive } from 'vue'
import SourceView from './components/SourceView.vue'

interface SourceTab {
  id: symbol
  name: string
  source: string
}

const tabs = reactive<Map<symbol, SourceTab>>(new Map())
let nextTabNumber = 1

function addTab() {
  const tabId = Symbol('source-tab')
  tabs.set(tabId, {
    id: tabId,
    name: `source${nextTabNumber}`,
    source: `// file ${nextTabNumber}\n\npackage main`,
  })
  nextTabNumber++
}

addTab()

onMounted(() => {
  getAvailableCompilers()
})

function onCloseTab(id: symbol) {
  tabs.delete(id)
}

function onTabRenamed(id: symbol, name: string) {
  let tab = tabs.get(id)
  if (tab) {
    tab.name = name
  }
}

async function getAvailableCompilers() {
  try {
    State.compilers = await API.listCompilers()
    if (State.compilers.length > 0) State.defaultCompiler = State.compilers[0].name
  } catch (e) {
    State.appendError('cannot get compilers')
  }
}
</script>

<template>
  <div class="root" :class="`theme-${State.theme}`">
    <MenuBar></MenuBar>

    <GoceTabs
      class="source-tabs"
      closable
      renameable
      newTabButton
      @newTabClicked="addTab"
      @closeTabClicked="onCloseTab"
      @tabRenamed="onTabRenamed"
    >
      <GoceTab v-for="[id, tab] in tabs.entries()" :key="id" :title="tab.name">
        <SourceView class="source-view" :value="tab.source"></SourceView>
      </GoceTab>
    </GoceTabs>

    <StatusBar></StatusBar>
  </div>
</template>

<style lang="scss">
@use '@/assets/themes/theme.scss';

.root {
  height: 100vh;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: theme.$backgroundColor;
  color-scheme: theme.$colorScheme;

  .source-tabs {
    margin-top: 0.5rem;
    flex: 1;
  }

  .source-view {
    height: 100%;
  }
}
</style>
