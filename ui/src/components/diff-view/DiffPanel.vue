<script setup lang="ts">
import DropDown from '@/components/ui/DropDown.vue'
import GoceButton from '@/components/ui/GoceButton.vue'
import GoceCheckbox from '@/components/ui/GoceCheckbox.vue'

import type { DiffSettings, SourceTab } from '@/tab'

import { computed, onMounted, reactive, ref } from 'vue'

const props = defineProps<{
  tabs: Map<symbol, SourceTab>
  settings: DiffSettings
}>()

const emit = defineEmits<{
  (e: 'update:settings', value: DiffSettings): void
}>()

const selected = reactive({
  original: -1,
  modified: -1,
})

const tabs = computed(() => Array.from(props.tabs.values()))
const tabIds = computed(() => tabs.value.map((tab) => tab.id))
const tabNames = computed(() => tabs.value.map((tab) => tab.name))

const showInline = ref(props.settings.inline)

onMounted(() => {
  if (props.settings.original) {
    selected.original = tabIds.value.indexOf(props.settings.original!)
  }

  if (props.settings.modified) {
    selected.modified = tabIds.value.indexOf(props.settings.modified!)
  }

  if (props.tabs.size > 0) {
    if (selected.original < 0) {
      selected.original = 0
      if (props.tabs.size > 1 && selected.modified >= 0) {
        for (let i = 0; i < tabs.value.length; i++) {
          if (selected.modified != i) {
            selected.original = i
            break
          }
        }
      }
    }
    if (selected.modified < 0) {
      selected.modified = 0
      if (props.tabs.size > 1 && selected.original >= 0) {
        for (let i = 0; i < tabs.value.length; i++) {
          if (selected.original != i) {
            selected.modified = i
            break
          }
        }
      }
    }
  }

  updateSettings()
})

function updateSettings() {
  emit('update:settings', {
    original: tabIds.value[selected.original],
    modified: tabIds.value[selected.modified],
    inline: showInline.value,
  })
}

function swapSources() {
  ;[selected.original, selected.modified] = [selected.modified, selected.original]
  updateSettings()
}
</script>

<template>
  <div class="diff-panel">
    <div class="labeled-item">
      <label>Original:</label>
      <DropDown
        class="control dropdown"
        v-model="selected.original"
        @update:modelValue="updateSettings"
        :options="tabNames"
      ></DropDown>
    </div>

    <div class="item">
      <GoceButton @click="swapSources">
        <i class="codicon codicon-arrow-swap"></i>
      </GoceButton>
    </div>
    <div class="labeled-item">
      <label>Modified:</label>
      <DropDown
        class="control"
        v-model="selected.modified"
        @update:modelValue="updateSettings"
        :options="tabNames"
      ></DropDown>
    </div>

    <div class="labeled-item">
      <label>Style:</label>
      <GoceCheckbox v-model="showInline" @update:modelValue="updateSettings" class="control">
        Inline
      </GoceCheckbox>
    </div>
  </div>
</template>

<style scoped lang="scss">
@use '@/assets/themes/theme.scss';

$labelHeight: 0.7rem;
$controlHeight: 1.5rem;

.diff-panel {
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
