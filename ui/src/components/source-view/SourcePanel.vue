<script setup lang="ts">
import DropDown from '@/components/ui/DropDown.vue'
import GoceButton from '@/components/ui/GoceButton.vue'
import GoceCheckbox from '@/components/ui/GoceCheckbox.vue'

import type { SourceSettings } from '@/tab'
import State from '@/state'

import { computed, onMounted, reactive, ref } from 'vue'

const props = defineProps<{
  settings: SourceSettings
}>()

const settings = reactive(props.settings)
const selectedCompilerIndex = ref(0)
const selectedArchitectureLevel = ref(0)
const optimizations = ref(!props.settings.compilerOptions.disableOptimizations)
const inlining = ref(!props.settings.compilerOptions.disableInlining)

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
    if (c.name == settings.compiler.name) {
      selectedCompilerIndex.value = i
      break
    }
  }
  selectedArchitectureLevel.value = availableLevels.value.default
  if (settings.compilerOptions.architectureLevel) {
    selectedArchitectureLevel.value = availableLevels.value.values.indexOf(settings.compilerOptions.architectureLevel)
  }
  updateSettings()
})

interface ArchitectureLevels {
  default: number
  names: Array<string>
  values: Array<string>
}

const availableLevels = computed(() => {
  let levels: ArchitectureLevels = {
    names: [],
    values: [],
    default: 0,
  }
  let c = State.compilers[selectedCompilerIndex.value]
  switch (c.architecture) {
    case 'amd64':
      levels.names = ['x86-64-v1', 'x86-64-v2', 'x86-64-v3', 'x86-64-v4']
      levels.values = ['v1', 'v2', 'v3', 'v4']
      levels.default = 0
      break
    case 'ppc64':
      levels.names = ['power8', 'power9']
      levels.values = ['power8', 'power9']
      levels.default = 0
      break
    case '386':
      levels.names = ['softfloat', 'sse2']
      levels.values = ['softfloat', 'sse2']
      levels.default = 1
      break
    case 'arm':
      levels.names = ['softfloat', 'VFPv1/2', 'VFPv3']
      levels.values = ['5', '6', '7']
      levels.default = 1
      break
  }
  return levels
})

function selectCompiler(index: number) {
  selectedCompilerIndex.value = index
  selectedArchitectureLevel.value = availableLevels.value.default
  updateSettings()
}

function selectArchitectureLevel(index: number) {
  selectedArchitectureLevel.value = index
  updateSettings()
}

function updateCompilerOptions() {
  updateSettings()
}

function updateSettings() {
  settings.compiler = State.compilers[selectedCompilerIndex.value]
  settings.compilerOptions.architectureLevel = availableLevels.value.values[selectedArchitectureLevel.value]
  settings.compilerOptions.disableOptimizations = !optimizations.value
  settings.compilerOptions.disableInlining = !inlining.value
  emit('update:settings', settings)
}
</script>

<template>
  <div class="source-panel">
    <div class="labeled-item">
      <label>Compiler:</label>
      <DropDown
        class="control dropdown"
        :modelValue="selectedCompilerIndex"
        @update:modelValue="selectCompiler"
        :options="compilerNames"
      ></DropDown>
    </div>

    <div class="labeled-item" v-if="availableLevels.values.length > 0">
      <label>Architecture level:</label>
      <DropDown
        class="control"
        style="width: 8rem"
        :modelValue="selectedArchitectureLevel"
        @update:modelValue="selectArchitectureLevel"
        :options="availableLevels.names"
      ></DropDown>
    </div>

    <div class="item">
      <GoceCheckbox class="control" v-model="optimizations" @update:modelValue="updateCompilerOptions"
        >Optimizations</GoceCheckbox
      >
    </div>

    <div class="item">
      <GoceCheckbox class="control" v-model="inlining" @update:modelValue="updateCompilerOptions"
        >Inlining</GoceCheckbox
      >
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
