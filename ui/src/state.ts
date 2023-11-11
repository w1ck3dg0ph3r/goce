import type { CompilerInfo } from '@/services/api'

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

  compilers: Array<CompilerInfo> = new Array()
  compilerByName: Map<string, CompilerInfo> = new Map()
  defaultCompiler: number = 0

  errorMessages = ''

  sharedCodeLink: string | null = null

  constructor() {
    this._theme = localStorage.getItem('theme') || 'light'
  }

  setError(error: string) {
    this.errorMessages = error
  }

  appendError(error: string) {
    this.errorMessages += error + '\n'
  }

  clearErrors() {
    this.errorMessages = ''
  }
}

const state = reactive(new State())
export default state
