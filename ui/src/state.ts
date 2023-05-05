import { SourceMap } from './components/editor/sourcemap'

import { reactive } from 'vue'

export interface State {
  theme: 'light' | 'dark'
  selectedCompiler: string

  status: Status
  errorMessage: string

  sourceMap: SourceMap
}

export enum Status {
  Idle,
  Formatting,
  Compiling,
}

const State = reactive<State>({
  theme: 'light',
  selectedCompiler: '',

  status: Status.Idle,
  errorMessage: '',

  sourceMap: new SourceMap(),
})

export default State
