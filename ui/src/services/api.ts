import type { SourceSettings } from "@/tab"

export class API {
  readonly baseUrl: string = import.meta.env.VITE_APP_API_BASE_URL

  constructor(baseUrl?: string) {
    if (baseUrl) {
      this.baseUrl = baseUrl
    }
  }

  async listCompilers(): Promise<CompilerInfo[]> {
    const res = await fetch(`${this.baseUrl}/api/compilers`)
    if (!res.ok) {
      throw Error(await res.text())
    }
    return await res.json()
  }

  async formatCode(code: string, compilerName?: string): Promise<FormattedCode> {
    const res = await fetch(`${this.baseUrl}/api/format`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        name: compilerName,
        code: code,
      }),
    })
    if (!res.ok) {
      throw await res.text()
    }
    return await res.json()
  }

  async compileCode(
    code: string,
    compilerName: string,
    compilerOptions?: CompilerOptions
  ): Promise<CompilationResult> {
    const res = await fetch(`${this.baseUrl}/api/compile`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        name: compilerName,
        options: compilerOptions,
        code: code,
      }),
    })
    if (!res.ok) {
      throw await res.text()
    }
    return await res.json()
  }

  async shareCode(code: SharedCode): Promise<string> {
    const res = await fetch(`${this.baseUrl}/api/shared`, {
      method: 'POST',
      body: JSON.stringify(code),
    })
    if (!res.ok) {
      throw Error('cannot share code')
    }
    const shared = await res.json()
    return shared.id
  }

  async getSharedCode(id: string): Promise<SharedCode> {
    const res = await fetch(`${this.baseUrl}/api/shared/${id}`)
    if (!res.ok) {
      throw Error('cannot get shared code')
    }
    return await res.json()
  }
}

export interface CompilerInfo {
  name: string
  version: string
  platform: string
  architecture: string
}

export interface CompilerOptions {
  disableInlining: boolean
  disableOptimizations: boolean
  architectureLevel: string
}

export interface FormattedCode {
  code: string
  errors: string
}

export interface CompilationResult {
  buildFailed: boolean
  buildOutput: string
  assembly?: string
  mapping?: {
    source: number
    start: number
    end: number
  }[]
  inliningAnalysis?: {
    name: string
    location: FileLocation
    canInline: boolean
    reason: string
    cost: number
  }[]
  inlinedCalls?: {
    name: string
    location: FileLocation
    length: number
  }[]
  heapEscapes?: {
    name: string
    location: FileLocation
  }[]
}

interface FileLocation {
  l: number
  c: number
}

export interface SharedCodeTab {
  type: 'code'
  code: string
  settings: SourceSettings
}

export interface SharedDiffTab {
  type: 'diff'
  originalSourceName: string
  modifiedSourceName: string
  inline: boolean
}

export type SharedTab = {
  name: string
} & (SharedCodeTab | SharedDiffTab)

export type SharedCode = Array<SharedTab>

const api = new API()
export default api
