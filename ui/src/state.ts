import { SourceMap } from './components/editor/sourcemap'

import { reactive } from 'vue'

class State {
  theme: 'light' | 'dark' = 'light'
  selectedCompiler: string = ''

  status: Status = Status.Idle
  errorMessage: string = ''

  sourceMap: SourceMap = new SourceMap()

  setError(error: string) {
    this.errorMessage = error
  }

  appendError(error: string) {
    this.errorMessage += error + '\n'
  }

  clearErrors() {
    this.errorMessage = ''
  }
}

export enum Status {
  Idle,
  Formatting,
  Compiling,
}

const state = reactive(new State())
export default state
