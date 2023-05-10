<script lang="ts" setup>
import API, { type CompilerInfo } from '@/services/api'
import State from '@/state'

import { computed, onMounted, reactive } from 'vue'
import GoceLogo from './GoceLogo.vue'
import MenuButton from './MenuButton.vue'
import DropDown from './DropDown.vue'

import '@vscode/codicons/dist/codicon.css'

defineEmits<{
  (event: 'format'): void
  (event: 'compile'): void
}>()

const state = reactive({
  compilers: Array<CompilerInfo>(),
})

onMounted(async () => {
  try {
    state.compilers = await API.listCompilers()
    if (state.compilers.length > 0) State.selectedCompiler = state.compilers[0].name
  } catch (e) {
    State.appendError('cannot get compilers')
  }
})

const compilerOptions = computed(() => {
  return state.compilers.map((c) => ({
    value: c.name,
    text: c.name,
  }))
})

function switchTheme() {
  State.theme = State.theme == 'light' ? 'dark' : 'light'
}
</script>

<template>
  <div id="menu">
    <GoceLogo class="logo"></GoceLogo>
    <div class="spacer"></div>
    <MenuButton @click="switchTheme">
      <i
        class="codicon codicon-color-mode"
        :style="{ transform: State.theme == 'dark' ? `rotate(180deg)` : undefined }"
      ></i>
      <span>{{ State.theme == 'light' ? 'Dark' : 'Light' }}</span>
    </MenuButton>
    <DropDown v-model="State.selectedCompiler" :options="compilerOptions"></DropDown>
    <MenuButton @click="$emit('format')">
      <i class="codicon codicon-json"></i>
      <span>Format</span>
    </MenuButton>
    <MenuButton @click="$emit('compile')">
      <i class="codicon codicon-play"></i>
      <span>Compile</span>
    </MenuButton>
  </div>
</template>

<style lang="scss" scoped>
@use '@/assets/themes/theme.scss';

$height: 1.5rem;

#menu {
  padding: 0.5rem;
  display: flex;
  gap: 1rem;

  .logo {
    height: $height;
  }

  .spacer {
    flex: 1;
  }
}
</style>
