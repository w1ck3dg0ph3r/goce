<script lang="ts" setup>
import MenuBar from '@/components/menubar/MenuBar.vue'
import CodeEditor from '@/components/CodeEditor.vue'
import AsmView from '@/components/AsmView.vue'
import OutputPane from '@/components/outputpane/OutputPane.vue'
import StatusBar from '@/components/statusbar/StatusBar.vue'
import { Panel, Splitter } from '@/components/ui/PanelSplitter.vue'
import LoadingIndicator from '@/components/ui/LoadingIndicator.vue'

import API from '@/services/api'
import bus from '@/services/bus'
import State, { Status } from '@/state'

import { onMounted, ref, watch } from 'vue'

const $codeEditor = ref<InstanceType<typeof CodeEditor> | null>(null)

onMounted(async () => {
  await getAvailableCompilers()
  State.sourceMap.init()
  await loadStartupCode()

  bus.on('formatCode', formatCode)
  bus.on('compileCode', compileCode)
  bus.on('shareCode', shareCode)

  useRecompileOnCompilerChange()
})

async function formatCode() {
  if (State.status == Status.Formatting) return
  let code = $codeEditor.value?.getCode()
  if (!code) return

  State.status = Status.Formatting
  try {
    let res = await API.formatCode(code, State.selectedCompiler)
    if (res.code !== '') {
      $codeEditor.value?.setCode(res.code, true)
    }
    if (res.errors) {
      State.setError(res.errors)
    } else {
      State.clearErrors()
    }
  } catch (e) {
    State.appendError('cannot format code')
  } finally {
    State.status = Status.Idle
  }
}

async function compileCode() {
  if (State.status == Status.Compiling) return
  let code = $codeEditor.value?.getCode()
  if (!code) return

  State.status = Status.Compiling
  State.buildOutput = ''
  try {
    let compiled = await API.compileCode(code, State.selectedCompiler)
    State.buildOutput = compiled.buildOutput
    if (compiled.buildFailed) {
      State.status = Status.Error
      State.sourceMap.assembly = ''
      return
    }
    State.sourceMap.update(compiled)
    State.status = Status.Idle
  } catch (e) {
    State.appendError('cannot compile code')
    State.status = Status.Error
  }
}

async function shareCode() {
  const code = $codeEditor.value?.getCode()
  if (!code) return
  try {
    const id = await API.shareCode(code)
    State.sharedCodeLink = `${import.meta.env.VITE_APP_BASE_URL}/${id}`
  } catch (e) {
    State.appendError('cannot share code')
  }
}

async function loadStartupCode() {
  let id = getSharedId()
  if (id) {
    try {
      let code = await API.getSharedCode(id)
      $codeEditor.value?.setCode(code, false)
    } catch (e) {
      State.appendError((e as Error).message)
    }
  } else {
    $codeEditor.value?.setCode(defaultCode, false)
  }
}

function getSharedId(): string {
  const path = document.location.pathname
  const matches = /\/([A-Za-z0-9]+)/.exec(path)
  if (!matches) return ''
  return matches[1]
}

async function getAvailableCompilers() {
  try {
    State.compilers = await API.listCompilers()
    if (State.compilers.length > 0) State.selectedCompiler = State.compilers[0].name
  } catch (e) {
    State.appendError('cannot get compilers')
  }
}

function useRecompileOnCompilerChange() {
  return watch(
    () => State.selectedCompiler,
    () => {
      compileCode()
    }
  )
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

    <Splitter horizontal class="main">
      <Panel :min-size="15">
        <Splitter>
          <Panel>
            <CodeEditor ref="$codeEditor" @change="compileCode"></CodeEditor>
          </Panel>
          <Panel class="asm-view">
            <AsmView></AsmView>
            <LoadingIndicator v-if="State.status == Status.Compiling"></LoadingIndicator>
          </Panel>
        </Splitter>
      </Panel>

      <Panel v-if="State.bottomPanelVisible" :min-size="15" :size="25">
        <OutputPane></OutputPane>
      </Panel>
    </Splitter>

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

  .main {
    flex: 1;

    .asm-view {
      position: relative;
    }
  }
}
</style>
