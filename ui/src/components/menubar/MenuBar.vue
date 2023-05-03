<script lang="ts" setup>
import API, { type CompilerInfo } from '@/services/api'
import State from '@/state'

import { onMounted, reactive } from 'vue'
import MenuButton from './MenuButton.vue'

import '@vscode/codicons/dist/codicon.css'

const state = reactive({
  compilers: Array<CompilerInfo>(),
})

onMounted(async () => {
  state.compilers = await API.listCompilers()
  if (state.compilers.length > 0) State.selectedCompiler = state.compilers[0].name
})

function switchTheme() {
  State.theme = State.theme == 'light' ? 'dark' : 'light'
}
</script>

<template>
  <div id="menu">
    <div class="logo">
      <img src="@/assets/logo.svg" />
    </div>
    <div class="spacer"></div>
    <MenuButton @click="switchTheme">
      <i class="codicon codicon-color-mode"></i>
      <span>{{ State.theme == 'light' ? 'Dark' : 'Light' }}</span>
    </MenuButton>
    <!-- <div class="text">Compiler:</div> -->
    <select v-model="State.selectedCompiler">
      <option v-for="c of state.compilers" :key="c.name" :value="c.name">{{ c.name }}</option>
    </select>
    <MenuButton @click="$emit('format')">
      <i class="codicon codicon-json"></i>
      <span>Format</span>
    </MenuButton>
    <MenuButton @click="$emit('build')">
      <i class="codicon codicon-play"></i>
      <span>Compile</span>
    </MenuButton>
  </div>
</template>

<style lang="scss" scoped>
$height: 1.5rem;

#menu {
  padding: 0.5rem;
  display: flex;
  gap: 1rem;

  .logo {
    > img {
      height: $height;
    }
  }

  .text {
    line-height: $height;
    font-size: 1rem;
  }

  .spacer {
    flex: 1;
  }
}
</style>
