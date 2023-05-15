<script lang="ts" setup>
import State from '@/state'
import GoceTabs from '@/components/ui/GoceTabs.vue'
import GoceTab from '@/components/ui/GoceTab.vue'

import BuildOutput from './BuildOutput.vue'
</script>

<template>
  <div class="output-pane">
    <GoceTabs class="tabs">
      <GoceTab title="Build Output" class="tab">
        <BuildOutput :value="State.buildOutput"></BuildOutput>
      </GoceTab>
    </GoceTabs>
  </div>
</template>

<style scoped lang="scss">
@use '@/assets/themes/theme.scss';

.output-pane {
  display: flex;

  .tabs {
    flex: 1;
    .tab {
      width: 100%;
      height: 100%;
    }
  }
}
</style>
