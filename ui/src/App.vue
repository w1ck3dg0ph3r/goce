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

import { computed, onMounted, reactive, ref, type Ref, nextTick } from 'vue'

const tabs = reactive<Array<Tab>>([])
const tabsById = reactive<Map<symbol, Tab>>(new Map())
const $tabs: Ref<InstanceType<typeof GoceTabs> | null> = ref(null)
let activeTabId: symbol
let nextSourceTabNumber = 1
let nextDiffTabNumber = 1

const sourceTabs = computed(() => {
  const m = new Map<symbol, SourceTab>()
  for (const tab of tabs.values()) {
    if (tab instanceof SourceTab) {
      m.set(tab.id, tab)
    }
  }
  return m
})

function addSourceTab() {
  const id = Symbol('tab')
  const name = `source${nextSourceTabNumber}`
  const tab = new SourceTab(id, name, defaultCode, {
    compiler: State.defaultCompiler,
    compilerOptions: {
      architectureLevel: '',
      disableOptimizations: false,
      disableInlining: false,
    },
  })
  tabs.push(tab)
  tabsById.set(id, tab)
  nextSourceTabNumber++
  return tab
}

function addDiffTab(originalId?: symbol, modifiedId?: symbol, inline?: boolean) {
  const id = Symbol('tab')
  const name = `diff${nextDiffTabNumber}`
  const tab = new DiffTab(id, name, {
    original: originalId,
    modified: modifiedId,
    inline: inline || false,
  })
  tabs.push(tab)
  tabsById.set(id, tab)
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
    const sharedId = document.location.pathname.substring(1)
    if (sharedId.length == 0) return false
    shared = await API.getSharedCode(sharedId)
    if (!shared || !shared.tabs || shared.tabs.length == 0) return false
  } catch (e) {
    State.appendError('cannot load shared code')
    return false
  }

  const diffTabs = new Array<SharedDiffTab>()

  for (const sharedTab of shared.tabs) {
    switch (sharedTab.type) {
      case 'code': {
        const id = Symbol('tab')
        const tab = new SourceTab(id, sharedTab.name, sharedTab.code, sharedTab.settings)
        if (!tab.settings.compiler || !isCompilerAvailable(tab.settings.compiler.name)) {
          tab.settings.compiler = State.defaultCompiler
          tab.settings.compilerOptions = {
            architectureLevel: '',
            disableInlining: false,
            disableOptimizations: false,
          }
        }
        tabs.push(tab)
        tabsById.set(tab.id, tab)
        break
      }
      case 'diff': {
        const id = Symbol('tab')
        const tab = new DiffTab(id, sharedTab.name, { inline: sharedTab.inline })
        tabs.push(tab)
        tabsById.set(id, tab)
        diffTabs.push(sharedTab)
        break
      }
    }
  }

  for (const [i, sharedTab] of shared.tabs.entries()) {
    if (sharedTab.type == 'diff') {
      const diffTab = tabs[i] as DiffTab
      diffTab.settings.original =
        sharedTab.originalSource >= 0 ? tabs[sharedTab.originalSource].id : undefined
      diffTab.settings.modified =
        sharedTab.modifiedSource >= 0 ? tabs[sharedTab.modifiedSource].id : undefined
    }
  }

  if (shared.activeTab >= 0 && shared.activeTab < tabs.length) {
    nextTick(() => {
      if ($tabs.value) {
        $tabs.value.selectTab(tabs[shared.activeTab].id)
      }
    })
  }

  return true
}

function isCompilerAvailable(compilerName: string): boolean {
  return State.compilerByName.has(compilerName)
}

bus.on('shareCode', async () => {
  const shared: SharedCode = {
    tabs: new Array<SharedTab>(),
    activeTab: -1,
  }
  const tabIdxById = new Map<symbol, number>()
  for (const [i, tab] of tabs.entries()) {
    tabIdxById.set(tab.id, i)
    if (tab instanceof SourceTab) {
      shared.tabs.push({
        name: tab.name,
        type: 'code',
        code: tab.code,
        settings: tab.settings,
      })
    } else if (tab instanceof DiffTab) {
      shared.tabs.push({
        name: tab.name,
        type: 'diff',
        originalSource: -1,
        modifiedSource: -1,
        inline: tab.settings.inline,
      })
    }
    if (tab.id == activeTabId) {
      shared.activeTab = i
    }
  }
  for (const [i, tab] of tabs.entries()) {
    if (tab instanceof DiffTab && shared.tabs[i]) {
      const diffTab = shared.tabs[i] as SharedDiffTab
      if (tab.settings.original) diffTab.originalSource = tabIdxById.get(tab.settings.original)!
      if (tab.settings.modified) diffTab.modifiedSource = tabIdxById.get(tab.settings.modified)!
    }
  }
  const link = await API.shareCode(shared)
  State.sharedCodeLink = `${import.meta.env.VITE_APP_BASE_URL}/${link}`
})

function onCloseTab(id: symbol) {
  const tab = tabsById.get(id)
  if (!tab) return
  tabs.splice(tabs.indexOf(tab), 1)
  tabsById.delete(id)
}

function onTabRenamed(id: symbol, name: string) {
  const tab = tabsById.get(id)
  if (!tab) return
  tab.name = name
}

function onTabsSwapped(i: number, j: number) {
  ;[tabs[i], tabs[j]] = [tabs[j], tabs[i]]
}

async function getAvailableCompilers() {
  try {
    State.compilers = await API.listCompilers()
    for (const c of State.compilers) State.compilerByName.set(c.name, c)
  } catch (e) {
    State.appendError('cannot get compilers')
  }
}

const defaultCode = `\
package main

import (
	"fmt"
	"math"
	"os"

	"github.com/shopspring/decimal"
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
	d := decimal.New(42, -1)
	fmt.Println(d.String())
	res := fibonacci(3)
	fmt.Println(res)
	fmt.Println(sqrt(float32(res)))
	fmt.Println(square(res))
	os.Exit(s[42])
}

var s = make([]int, 100)
`

function tabIcon(tab: Tab): string | undefined {
  if (tab instanceof SourceTab) return 'file-code'
  if (tab instanceof DiffTab) return 'diff'
}
</script>

<template>
  <div class="root" :class="`theme-${State.theme}`">
    <MenuBar />

    <GoceTabs
      ref="$tabs"
      class="source-tabs"
      closable
      renameable
      new-tab-button
      @new-tab-clicked="addSourceTab"
      @close-tab-clicked="onCloseTab"
      @tab-selected="activeTabId = $event"
      @tab-renamed="onTabRenamed"
      @tabs-swapped="onTabsSwapped"
    >
      <GoceTab
        v-for="(tab, idx) of tabs"
        :key="tab.id"
        :tab-id="tab.id"
        :title="tab.name"
        :icon="tabIcon(tab)"
        :order="idx"
      >
        <SourceView
          v-if="tab instanceof SourceTab"
          v-model:code="tab.code"
          v-model:settings="tab.settings"
          v-model:source-map="tab.sourceMap"
          class="source-view"
          @diff="addDiffTab(undefined, tab.id)"
        />
        <DiffView
          v-if="tab instanceof DiffTab"
          v-model:settings="tab.settings"
          :tabs="sourceTabs"
          class="diff-view"
        />
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
