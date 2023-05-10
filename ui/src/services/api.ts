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

  async compile(code: string, compilerName?: string): Promise<CompilationResult> {
    const res = await fetch(`${this.baseUrl}/api/compile`, {
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

  async shareCode(code: string): Promise<string> {
    const res = await fetch(`${this.baseUrl}/api/shared`, {
      method: 'POST',
      body: code,
    })
    if (!res.ok) {
      throw Error('cannot share code')
    }
    const shared = await res.json()
    return shared.id
  }

  async getSharedCode(id: string): Promise<string> {
    const res = await fetch(`${this.baseUrl}/api/shared/${id}`)
    if (!res.ok) {
      throw Error('cannot get shared code')
    }
    return await res.text()
  }
}

export interface CompilerInfo {
  name: string
  version: string
  platform: string
}

export interface FormattedCode {
  code: string
  errors: string
}

export interface CompilationResult {
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
  errors?: string
}

interface FileLocation {
  l: number
  c: number
}

const api = new API()
export default api
