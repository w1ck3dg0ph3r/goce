<script lang="ts" setup>
import CodeEditor from '@/components/CodeEditor.vue'
import AsmView from '@/components/AsmView.vue'
import OutputPane from '@/components/outputpane/OutputPane.vue'
import { Panel, Splitter } from '@/components/ui/PanelSplitter.vue'
import LoadingIndicator from '@/components/ui/LoadingIndicator.vue'
import DropDown from '@/components/ui/DropDown.vue'
import GoceButton from './ui/GoceButton.vue'

import API from '@/services/api'
import State, { Status } from '@/state'
import { SourceMap } from './editor/sourcemap'

import { computed, onMounted, reactive, ref, watch } from 'vue'
import { debounce } from 'lodash'

const props = defineProps<{
  code?: string
}>()

const emit = defineEmits<{
  (e: 'update:code', code: string): void
}>()

const $codeEditor = ref<InstanceType<typeof CodeEditor> | null>(null)
const $asmView = ref<InstanceType<typeof AsmView> | null>(null)

const state = reactive({
  compiler: State.defaultCompiler,

  buildOutput: '',
  sourceMap: new SourceMap(),
})

watch(
  () => State.compilers,
  () => {
    if (state.compiler == '' && State.compilers.length > 0) {
      state.compiler = State.compilers[0].name
    }
  }
)

const compilerOptions = computed(() => {
  return State.compilers.map((c) => ({
    value: c.name,
    text: c.name,
  }))
})

async function formatCode() {
  if (State.status == Status.Formatting) return
  if (!props.code) return

  State.status = Status.Formatting
  try {
    let res = await API.formatCode(props.code, state.compiler)
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
    State.status = Status.Idle
  }
}

async function compileCode() {
  if (State.status == Status.Compiling) return
  if (!props.code) return

  State.status = Status.Compiling
  state.buildOutput = ''
  try {
    let compiled = await API.compileCode(props.code, state.compiler)
    state.buildOutput = compiled.buildOutput
    if (compiled.buildFailed) {
      State.status = Status.Error
      state.sourceMap.assembly = ''
      return
    }
    state.sourceMap.update(compiled)
    State.status = Status.Idle
  } catch (e) {
    State.appendError('cannot compile code')
    State.status = Status.Error
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

onMounted(() => {
  if (state.compiler != '' && props.code != '') {
    compileCode()
  }
  watch(
    () => state.compiler,
    () => {
      compileCode()
    }
  )
})
</script>

<template>
  <div class="source-view">
    <div class="menu">
      <DropDown v-model="state.compiler" :options="compilerOptions"></DropDown>
      <div class="spacer"></div>
      <GoceButton @click="formatCode">
        <i class="codicon codicon-json"></i>
        <span>Format</span>
      </GoceButton>
      <GoceButton @click="compileCode">
        <i class="codicon codicon-play"></i>
        <span>Compile</span>
      </GoceButton>
    </div>

    <Splitter horizontal class="main">
      <Panel :min-size="15">
        <Splitter>
          <Panel>
            <CodeEditor
              ref="$codeEditor"
              :code="props.code"
              @update:code="updateCode"
              :sourceMap="state.sourceMap"
              @formatCode="formatCode"
              @lineHovered="state.sourceMap.highlightFromSource($event)"
              @revealAssembly="revealAssembly"
            ></CodeEditor>
          </Panel>
          <Panel class="asm-view">
            <AsmView
              ref="$asmView"
              :sourceMap="state.sourceMap"
              @lineHovered="state.sourceMap.highlightFromAssembly($event)"
              @revealSource="revealSource"
            ></AsmView>
            <LoadingIndicator v-if="State.status == Status.Compiling"></LoadingIndicator>
          </Panel>
        </Splitter>
      </Panel>

      <Panel v-if="State.bottomPanelVisible" :min-size="15" :size="25">
        <OutputPane :buildOutput="state.buildOutput" @jumpToSource="jumpToSource"></OutputPane>
      </Panel>
    </Splitter>
  </div>
</template>

<style scoped lang="scss">
@use '@/assets/themes/theme.scss';

.source-view {
  display: flex;
  flex-direction: column;
  background-color: theme.$backgroundColor;
  color-scheme: theme.$colorScheme;

  .menu {
    height: 2.5rem;
    padding: 0.5rem;
    display: flex;
    gap: 0.5rem;
  }

  .spacer {
    flex: 1;
  }

  .asm-view {
    position: relative;
  }
}
</style>
