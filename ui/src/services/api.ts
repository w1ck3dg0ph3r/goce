export class API {
  readonly baseUrl: string = 'http://127.0.0.1:9000'

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

  async formatCode(code: string): Promise<FormattedCode> {
    const res = await fetch(`${this.baseUrl}/api/format`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        code: code,
      }),
    })
    if (!res.ok) {
      throw await res.text()
    }
    return await res.json()
  }

  async compile(code: string): Promise<CompilationResult> {
    const res = await fetch(`${this.baseUrl}/api/compile`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        code: code,
      }),
    })
    if (!res.ok) {
      throw await res.text()
    }
    return await res.json()
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
