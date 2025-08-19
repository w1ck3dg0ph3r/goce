<script lang="ts" setup>
import GoceTabs from '@/components/ui/GoceTabs.vue'
import GoceTab from '@/components/ui/GoceTab.vue'

import BuildOutput from './BuildOutput.vue'

const props = defineProps<{
  buildOutput?: string
}>()

const emit = defineEmits<{
  (e: 'jumpToSource', line: number, column?: number): void
}>()

function jumpToSource(line: number, column?: number) {
  emit('jumpToSource', line, column)
}

const tabIds = {
  buildOutput: Symbol('tab'),
}
</script>

<template>
  <div class="output-pane">
    <GoceTabs class="tabs">
      <GoceTab
        id="build-output"
        title="Build Output"
        icon="server-process"
        :order="0"
        :tab-id="tabIds.buildOutput"
        class="tab"
      >
        <BuildOutput :value="props.buildOutput" @jump-to-source="jumpToSource" />
      </GoceTab>
    </GoceTabs>
  </div>
</template>

<style scoped lang="scss">
@use '@/assets/themes/theme.scss';

.output-pane {
  flex: 1;

  .tabs {
    height: 100%;
  }
}
</style>
