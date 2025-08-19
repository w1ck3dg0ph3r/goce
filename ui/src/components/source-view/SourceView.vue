<script lang="ts" setup>
import OutputPane from '@/components/outputpane/OutputPane.vue'
import StatusBar, { Status } from '@/components/statusbar/StatusBar.vue'
import { Panel, Splitter } from '@/components/ui/PanelSplitter.vue'
import LoadingIndicator from '@/components/ui/LoadingIndicator.vue'

import type { SourceSettings } from '@/tab'
import SourcePanel from './SourcePanel.vue'
import CodeEditor from './CodeEditor.vue'
import AsmView from './AsmView.vue'

import API from '@/services/api'
import State from '@/state'
import type { SourceMap } from '@/components/editor/sourcemap'

import { reactive, ref } from 'vue'
import { debounce } from 'lodash'

const props = defineProps<{
  code: string
  settings: SourceSettings
  sourceMap: SourceMap
}>()

const emit = defineEmits<{
  (e: 'update:code', code: string): void
  (e: 'update:settings', settings: SourceSettings): void
  (e: 'diff'): void
}>()

const $codeEditor = ref<InstanceType<typeof CodeEditor> | null>(null)
const $asmView = ref<InstanceType<typeof AsmView> | null>(null)

const state = reactive({
  settings: props.settings,
  buildOutput: '',
  sourceMap: props.sourceMap,

  status: Status.Idle,
  bottomPanelVisible: true,
  cursorPosition: {
    lineNumber: 1,
    column: 1,
  },
})

async function formatCode() {
  if (state.status == Status.Formatting) return
  if (!props.code) return
  if (!props.settings.compiler) return

  state.status = Status.Formatting
  try {
    const res = await API.formatCode(props.code, props.settings.compiler.name)
    if (res.code !== '') {
      updateCode(res.code)
    }
    if (res.errors) {
      State.setError(res.errors)
    } else {
      State.clearErrors()
    }
  } catch (e) {
    State.appendError('cannot format code')
  } finally {
    state.status = Status.Idle
  }
}

async function compileCode() {
  if (state.status == Status.Compiling) return
  if (!props.code) return
  if (!props.settings.compiler) return

  state.status = Status.Compiling
  state.buildOutput = ''
  try {
    const compiled = await API.compileCode(
      props.code,
      props.settings.compiler.name,
      props.settings.compilerOptions
    )
    state.buildOutput = compiled.buildOutput
    state.sourceMap.update(compiled)
    state.status = compiled.buildFailed ? Status.Error : Status.Idle
  } catch (e) {
    State.appendError('cannot compile code')
    state.status = Status.Error
  }
}

function revealAssembly(sourceLineNumber: number) {
  const asmRanges = state.sourceMap.map.get(sourceLineNumber)?.ranges
  if (asmRanges && asmRanges.length > 0) {
    const firstAsmLine = asmRanges[0].start
    $asmView.value?.revealLine(firstAsmLine)
  }
}

function revealSource(assemblyLineNumber: number) {
  const sourceLine = state.sourceMap.reverseMap.get(assemblyLineNumber)
  if (sourceLine) {
    $codeEditor.value?.revealLine(sourceLine)
  }
}

function jumpToSource(line: number, column?: number) {
  $codeEditor.value?.jumpToLocation(line, column)
}

const debouncedCompileCode = debounce(compileCode, 1000)

function updateCode(code: string) {
  emit('update:code', code)
  debouncedCompileCode()
}

function applySettings() {
  emit('update:settings', state.settings)
  compileCode()
}
</script>

<template>
  <div class="source-view">
    <SourcePanel
      v-model:settings="state.settings"
      @update:settings="applySettings"
      @diff="emit('diff')"
      @format="formatCode"
      @compile="compileCode"
    />

    <Splitter horizontal class="main">
      <Panel :min-size="15">
        <Splitter>
          <Panel>
            <CodeEditor
              ref="$codeEditor"
              :code="props.code"
              :source-map="state.sourceMap"
              @update:code="updateCode"
              @format-code="formatCode"
              @cursor-moved="state.cursorPosition = $event"
              @line-hovered="state.sourceMap.highlightFromSource($event)"
              @reveal-assembly="revealAssembly"
            />
          </Panel>
          <Panel class="asm-view">
            <AsmView
              ref="$asmView"
              :source-map="state.sourceMap"
              @line-hovered="state.sourceMap.highlightFromAssembly($event)"
              @reveal-source="revealSource"
            />
            <LoadingIndicator v-if="state.status == Status.Compiling" />
          </Panel>
        </Splitter>
      </Panel>

      <Panel v-if="state.bottomPanelVisible" :min-size="15" :size="25">
        <OutputPane :build-output="state.buildOutput" @jump-to-source="jumpToSource" />
      </Panel>
    </Splitter>
    <StatusBar
      :state="{
        status: state.status,
        bottomPaneVisible: state.bottomPanelVisible,
        cursorPosition: state.cursorPosition,
      }"
      @toggle-bottom-panel="state.bottomPanelVisible = !state.bottomPanelVisible"
    />
  </div>
</template>

<style scoped lang="scss">
@use '@/assets/themes/theme.scss';

.source-view {
  display: flex;
  flex-direction: column;
  background-color: theme.$backgroundColor;
  color-scheme: theme.$colorScheme;

  .asm-view {
    position: relative;
  }
}
</style>
