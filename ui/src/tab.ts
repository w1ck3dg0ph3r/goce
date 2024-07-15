import { SourceMap } from '@/components/editor/sourcemap'
import type { CompilerInfo, CompilerOptions } from '@/services/api'

export class Tab {
  id: symbol
  name: string
  constructor(id: symbol, name: string) {
    this.id = id
    this.name = name
  }
}

export class SourceTab extends Tab {
  code: string
  settings: SourceSettings
  sourceMap: SourceMap

  constructor(id: symbol, name: string, code: string, settings: SourceSettings) {
    super(id, name)
    this.code = code
    this.settings = settings
    this.sourceMap = new SourceMap()
  }
}

export interface SourceSettings {
  compiler: CompilerInfo
  compilerOptions: CompilerOptions
}

export class DiffTab extends Tab {
  settings: DiffSettings

  constructor(id: symbol, name: string, settings?: DiffSettings) {
    super(id, name)
    this.settings = {
      original: settings?.original,
      modified: settings?.modified,
      inline: settings?.inline || false,
    }
  }
}

export interface DiffSettings {
  original?: symbol
  modified?: symbol
  inline: boolean
}
