import { SourceMap } from '@/components/editor/sourcemap'

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
  compiler: string
}

export class DiffTab extends Tab {
  settings: DiffSettings

  constructor(id: symbol, name: string, settings?: DiffSettings) {
    super(id, name)
    this.settings = {
      original: settings?.original || Symbol(),
      modified: settings?.modified || Symbol(),
      inline: settings?.inline || false,
    }
  }
}

export interface DiffSettings {
  original: symbol
  modified: symbol
  inline: boolean
}
