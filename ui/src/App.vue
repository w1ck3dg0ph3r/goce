<script lang="ts" setup>
import State from '@/state'
import API from '@/services/api'
import bus from '@/services/bus'

import MenuBar from '@/components/menubar/MenuBar.vue'
import StatusBar from '@/components/statusbar/StatusBar.vue'
import GoceTabs from '@/components/ui/GoceTabs.vue'
import GoceTab from '@/components/ui/GoceTab.vue'
import SourceView from '@/components/SourceView.vue'

import { onMounted, reactive } from 'vue'

interface SourceTab {
  id: symbol
  name: string
  code: string
}

const tabs = reactive<Map<symbol, SourceTab>>(new Map())
let nextTabNumber = 1

function addTab() {
  const tabId = Symbol('source-tab')
  tabs.set(tabId, {
    id: tabId,
    name: `source${nextTabNumber}`,
    code: defaultCode,
  })
  nextTabNumber++
}

onMounted(async () => {
  getAvailableCompilers()
  if (!(await loadSharedCode())) {
    addTab()
  }
})

async function loadSharedCode(): Promise<boolean> {
  try {
    let sharedId = document.location.pathname.substring(1)
    if (sharedId.length == 0) return false
    let shared = await API.getSharedCode(sharedId)
    if (!shared || shared.length == 0) return false
    for (let tab of shared) {
      let id = Symbol('source-tab')
      tabs.set(id, {
        id: id,
        name: tab.name,
        code: tab.code,
      })
    }
    return true
  } catch (e) {
    State.appendError('cannot load shared code')
    return false
  }
}

bus.on('shareCode', async () => {
  let shared = []
  for (let v of tabs.values()) {
    shared.push({
      name: v.name,
      code: v.code,
    })
  }
  let link = await API.shareCode(shared)
  State.sharedCodeLink = `${API.baseUrl}/${link}`
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

const defaultCode = `package main

import (
	"fmt"
	"math"
)

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

func square(n int) int {
	return n * n
}

func sqrt(x float32) float32 {
	return float32(math.Sqrt(float64(x)))
}

func main() {
	res := fibonacci(3)
	fmt.Println(res)
	fmt.Println(sqrt(float32(res)))
	fmt.Println(square(res))
}
`
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
        <SourceView class="source-view" v-model:code="tab.code"></SourceView>
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
