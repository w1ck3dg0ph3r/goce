<script setup lang="ts">
import DropDown from '@/components/ui/DropDown.vue'
import GoceButton from '@/components/ui/GoceButton.vue'
import type { SourceSettings } from '@/tab'

import State from '@/state'

import { computed, onMounted, reactive, ref } from 'vue'

const props = defineProps<{
  settings: SourceSettings
}>()

const settings = reactive(props.settings)
const selectedIndex = ref(0)

const compilerNames = computed(() => {
  return State.compilers.map((c) => c.name)
})

const emit = defineEmits<{
  (e: 'update:settings', settings: SourceSettings): void
  (e: 'diff'): void
  (e: 'format'): void
  (e: 'compile'): void
}>()

onMounted(() => {
  for (let [i, c] of State.compilers.entries()) {
    if (c.name == settings.compiler) {
      selectedIndex.value = i
      break
    }
  }
})

function selectCompiler(index: number) {
  selectedIndex.value = index
  settings.compiler = compilerNames.value[index]
  emit('update:settings', settings)
}

function updateOptions() {
  emit('update:settings', settings)
}
</script>

<template>
  <div class="source-panel">
    <div class="labeled-item">
      <label>Compiler:</label>
      <DropDown
        class="control dropdown"
        :modelValue="selectedIndex"
        @update:modelValue="selectCompiler"
        :options="compilerNames"
      ></DropDown>
    </div>

    <div class="labeled-item">
      <label>Options:</label>
      <input
        type="text"
        v-model="settings.options"
        @input="updateOptions"
        style="height: 1.5rem; width: 20rem"
      />
    </div>

    <div class="spacer"></div>

    <div class="item">
      <GoceButton @click="emit('diff')">
        <i class="codicon codicon-diff"></i>
        <span>Compare</span>
      </GoceButton>
    </div>
    <div class="item">
      <GoceButton @click="emit('format')">
        <i class="codicon codicon-json"></i>
        <span>Format</span>
      </GoceButton>
    </div>
    <div class="item">
      <GoceButton @click="emit('compile')">
        <i class="codicon codicon-play"></i>
        <span>Compile</span>
      </GoceButton>
    </div>
  </div>
</template>

<style scoped lang="scss">
@use '@/assets/themes/theme.scss';

$labelHeight: 0.7rem;
$controlHeight: 1.5rem;

.source-panel {
  padding: 0.5rem;
  display: flex;
  gap: 0.5rem;

  > .item {
    padding-top: $labelHeight;
    display: flex;
    flex-direction: column;
    > * {
      height: $controlHeight;
    }
  }

  > .labeled-item {
    display: flex;
    flex-direction: column;
    label {
      display: block;
      height: $labelHeight;
      font-size: $labelHeight;
      color: theme.$logoColor;
    }
    .control {
      height: $controlHeight;
    }
    .dropdown {
      width: 12rem;
    }
  }

  .spacer {
    flex: 1;
  }
}
</style>
