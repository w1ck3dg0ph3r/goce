import { SourceMap } from '@/components/editor/sourcemap'

import { Position } from 'monaco-editor'

import { reactive } from 'vue'

class State {
  private _theme: string
  get theme() {
    return this._theme
  }
  set theme(v: string) {
    this._theme = v
    localStorage.setItem('theme', v)
  }

  selectedCompiler: string = ''

  status: Status = Status.Idle
  errorMessage: string = ''

  sourceMap: SourceMap = new SourceMap()

  cursorPosition: Position = new Position(1, 1)

  constructor() {
    this._theme = localStorage.getItem('theme') || 'light'
  }

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
