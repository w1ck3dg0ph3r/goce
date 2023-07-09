<script lang="ts" setup>
import State from '@/state'

import GoceLogo from './GoceLogo.vue'
import ShareButton from './ShareButton.vue'
import GoceButton from '@/components/ui/GoceButton.vue'

import '@vscode/codicons/dist/codicon.css'

function switchTheme() {
  State.theme = State.theme == 'light' ? 'dark' : 'light'
}
</script>

<template>
  <div id="menu">
    <a href="/">
      <GoceLogo class="logo"></GoceLogo>
    </a>

    <GoceButton @click="switchTheme">
      <i
        class="codicon codicon-color-mode"
        :style="{ transform: State.theme == 'dark' ? `rotate(180deg)` : undefined }"
      ></i>
    </GoceButton>

    <ShareButton></ShareButton>

    <div class="spacer"></div>

    <a href="https://github.com/w1ck3dg0ph3r/goce" target="_blank">
      <GoceButton style="height: 100%">
        <i class="codicon codicon-github"></i>
        Source
      </GoceButton>
    </a>
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
