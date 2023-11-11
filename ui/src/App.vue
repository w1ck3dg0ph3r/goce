<script lang="ts" setup>
import State from '@/state'
import API, { type SharedCode, type SharedDiffTab, type SharedTab } from '@/services/api'
import bus from '@/services/bus'

import { Tab, SourceTab, DiffTab } from '@/tab'

import MenuBar from '@/components/menubar/MenuBar.vue'
import GoceTabs from '@/components/ui/GoceTabs.vue'
import GoceTab from '@/components/ui/GoceTab.vue'
import SourceView from '@/components/source-view/SourceView.vue'
import DiffView from '@/components/diff-view/DiffView.vue'

import { computed, onMounted, reactive } from 'vue'

const tabs = reactive<Map<symbol, Tab>>(new Map())
let nextSourceTabNumber = 1
let nextDiffTabNumber = 1

const sourceTabs = computed(() => {
  let m = new Map<symbol, SourceTab>()
  for (let tab of tabs.values()) {
    if (tab instanceof SourceTab) {
      m.set(tab.id, tab)
    }
  }
  return m
})

const sourceTabsByName = computed(() => {
  let m = new Map<string, SourceTab>()
  for (let [_, tab] of sourceTabs.value) {
    m.set(tab.name, tab)
  }
  return m
})

function addSourceTab() {
  const tabId = Symbol('source-tab')
  const tab = new SourceTab(tabId, `source${nextSourceTabNumber}`, defaultCode, {
    compilerName: State.defaultCompiler,
    compilerInfo: State.compilerByName.get(State.defaultCompiler)!,
    compilerOptions: {
      architectureLevel: '',
      disableOptimizations: false,
      disableInlining: false,
    },
  })
  tabs.set(tabId, tab)
  nextSourceTabNumber++
  return tab
}

function addDiffTab(originalId?: symbol, modifiedId?: symbol, inline?: boolean) {
  const tabId = Symbol('diff-tab')
  tabs.set(
    tabId,
    new DiffTab(tabId, `diff${nextDiffTabNumber}`, {
      original: originalId || Symbol(),
      modified: modifiedId || Symbol(),
      inline: inline || false,
    })
  )
  nextDiffTabNumber++
}

onMounted(async () => {
  await getAvailableCompilers()
  if (!(await loadSharedCode())) {
    addSourceTab()
  }
})

async function loadSharedCode(): Promise<boolean> {
  let shared: SharedCode
  try {
    let sharedId = document.location.pathname.substring(1)
    if (sharedId.length == 0) return false
    shared = await API.getSharedCode(sharedId)
    if (!shared || shared.length == 0) return false
  } catch (e) {
    State.appendError('cannot load shared code')
    return false
  }

  let diffTabs = new Array<SharedDiffTab>()

  for (let sharedTab of shared) {
    switch (sharedTab.type) {
      case 'code':
        let id = Symbol('source-tab')
        let tab = new SourceTab(id, sharedTab.name, sharedTab.code, sharedTab.settings)
        if (!isCompilerAvailable(tab.settings.compilerName)) {
          tab.settings.compilerName = State.defaultCompiler
        }
        tabs.set(id, tab)
        break
      case 'diff':
        diffTabs.push(sharedTab)
        break
    }
  }

  for (let sharedDiff of diffTabs) {
    let ot = sourceTabsByName.value.get(sharedDiff.originalSourceName)?.id
    let mt = sourceTabsByName.value.get(sharedDiff.modifiedSourceName)?.id
    addDiffTab(ot, mt, sharedDiff.inline)
  }

  return true
}

function isCompilerAvailable(compilerName: string): boolean {
  return compilerName in State.compilers
}

bus.on('shareCode', async () => {
  let shared = new Array<SharedTab>()
  for (let v of tabs.values()) {
    if (v instanceof SourceTab) {
      shared.push({
        name: v.name,
        type: 'code',
        code: v.code,
        settings: v.settings,
      })
    } else if (v instanceof DiffTab) {
      shared.push({
        name: v.name,
        type: 'diff',
        originalSourceName: tabs.get(v.settings.original)?.name || '',
        modifiedSourceName: tabs.get(v.settings.modified)?.name || '',
        inline: v.settings.inline,
      })
    }
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
    for (let c of State.compilers) State.compilerByName.set(c.name, c)
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
      @newTabClicked="addSourceTab"
      @closeTabClicked="onCloseTab"
      @tabRenamed="onTabRenamed"
    >
      <GoceTab v-for="[id, tab] in tabs.entries()" :key="id" :title="tab.name">
        <SourceView
          v-if="tab instanceof SourceTab"
          class="source-view"
          v-model:code="tab.code"
          v-model:settings="tab.settings"
          v-model:sourceMap="tab.sourceMap"
          @diff="addDiffTab(undefined, id)"
        ></SourceView>
        <DiffView
          v-if="tab instanceof DiffTab"
          :tabs="sourceTabs"
          v-model:settings="tab.settings"
          class="diff-view"
        ></DiffView>
      </GoceTab>
    </GoceTabs>
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

  .source-view,
  .diff-view {
    height: 100%;
  }
}
</style>
