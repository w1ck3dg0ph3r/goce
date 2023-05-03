import { reactive } from 'vue'

export interface State {
  theme: 'light' | 'dark'
  selectedCompiler: string

  status: Status
  errorMessage: string
}

export enum Status {
  Idle,
  Formatting,
  Compiling,
}

export default reactive<State>({
  theme: 'light',
  selectedCompiler: '',

  status: Status.Idle,
  errorMessage: '',
})

