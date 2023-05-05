<script lang="ts" setup>
import MenuBar from '@/components/menubar/MenuBar.vue'
import CodeEditor from '@/components/CodeEditor.vue'
import AsmView from '@/components/AsmView.vue'
import OutputPane from '@/components/OutputPane.vue'
import StatusBar from '@/components/statusbar/StatusBar.vue'

import API from '@/services/api'
import State, { Status } from '@/state'

import { onMounted, ref } from 'vue'

const $codeEditor = ref<InstanceType<typeof CodeEditor> | null>(null)

onMounted(() => {
  State.sourceMap.init()
  compileCode()
})

async function formatCode() {
  if (State.status != Status.Idle) return
  let code = $codeEditor.value?.getCode()
  if (!code) return

  State.status = Status.Formatting
  let res = await API.formatCode(code)
  if (res.code !== '') {
    $codeEditor.value?.setCode(res.code, true)
  }
  if (res.errors) {
    State.errorMessage = res.errors
  } else {
    State.errorMessage = ''
  }
  State.status = Status.Idle
}

async function compileCode() {
  if (State.status != Status.Idle) return
  let code = $codeEditor.value?.getCode()
  if (!code) return

  State.status = Status.Compiling
  let compiled = await API.compile(code)
  if (compiled.errors) {
    State.errorMessage = compiled.errors
    State.status = Status.Idle
    return
  } else {
    State.errorMessage = ''
  }
  State.sourceMap.update(compiled)
  State.status = Status.Idle
}

const defaultCode = `package main

import (
	"fmt"
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

func main() {
	res := fibonacci(3)
	fmt.Println(res)
	fmt.Println(square(res))
}
`
</script>

<template>
  <div class="root" :class="`theme-${State.theme}`">
    <MenuBar @format="formatCode" @build="compileCode"></MenuBar>
    <div class="split">
      <CodeEditor
        class="code"
        ref="$codeEditor"
        :defaultCode="defaultCode"
        @change="compileCode"
      ></CodeEditor>
      <AsmView class="assembly"></AsmView>
    </div>
    <div class="bottom">
      <OutputPane></OutputPane>
      <StatusBar></StatusBar>
    </div>
  </div>
</template>

<style lang="scss">
@use '@/reset.scss';
@use '@/assets/themes/theme.scss';

body {
  font-size: 14px;
}

.root {
  height: 100vh;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: theme.$backgroundColor;

  .split {
    flex: 1;
    display: flex;
    gap: 0.5rem;

    .code,
    .assembly {
      width: 50%;
    }
  }

  .bottom {
    flex: 0;
  }
}
</style>
