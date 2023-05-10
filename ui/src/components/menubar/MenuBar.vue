<script lang="ts" setup>
import bus from '@/services/bus'
import State from '@/state'

import { computed } from 'vue'
import GoceLogo from './GoceLogo.vue'
import MenuButton from './MenuButton.vue'
import ShareButton from './ShareButton.vue'
import DropDown from './DropDown.vue'

import '@vscode/codicons/dist/codicon.css'

const compilerOptions = computed(() => {
  return State.compilers.map((c) => ({
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
    <a href="/">
      <GoceLogo class="logo"></GoceLogo>
    </a>

    <MenuButton @click="switchTheme">
      <i
        class="codicon codicon-color-mode"
        :style="{ transform: State.theme == 'dark' ? `rotate(180deg)` : undefined }"
      ></i>
    </MenuButton>

    <ShareButton></ShareButton>

    <div class="spacer"></div>

    <DropDown v-model="State.selectedCompiler" :options="compilerOptions"></DropDown>

    <MenuButton @click="bus.emit('formatCode')">
      <i class="codicon codicon-json"></i>
      <span>Format</span>
    </MenuButton>

    <MenuButton @click="bus.emit('compileCode')">
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

  > a {
    border: none;
    text-decoration: none;
    .logo {
      display: block;
      height: $height;
      flex-shrink: 0;
    }
  }

  .spacer {
    flex: 1;
  }
}
</style>
