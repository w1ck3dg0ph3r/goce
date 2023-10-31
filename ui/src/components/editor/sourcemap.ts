import * as monaco from 'monaco-editor/esm/vs/editor/editor.api'

import type { CompilationResult } from '@/services/api'

import './sourcemap.scss'
import bus from '@/services/bus'

interface Mapping {
  color: string
  colorIdx: number
  ranges: {
    start: number
    end: number
  }[]
}

interface Assembly {
  code: string
  addresses: Array<string>
}

export class SourceMap {
  assembly: Assembly
  map: Map<number, Mapping>
  reverseMap: Map<number, number>

  sourceDecorations: monaco.editor.IModelDeltaDecoration[]
  assemblyDecorations: monaco.editor.IModelDeltaDecoration[]

  highlightedSource?: monaco.Range
  highlightedAssembly?: monaco.Range[]

  constructor() {
    this.assembly = { code: '', addresses: [] }
    this.map = new Map()
    this.reverseMap = new Map()
    this.sourceDecorations = new Array()
    this.assemblyDecorations = new Array()
  }

  init() {
    bus.on('sourceLineHovered', (ln) => this.highlightFromSource(ln))
    bus.on('assemblyLineHovered', (ln) => this.highlightFromAssembly(ln))
  }

  update(compiled: CompilationResult) {
    if (compiled.buildFailed) {
      this.assembly.code = ''
      this.assembly.addresses.length = 0
      this.assemblyDecorations.length = 0
      if (this.highlightedAssembly) this.highlightedAssembly.length = 0
      return
    }

    this.assembly = splitAssembly(compiled.assembly || '')
    this.map = new Map()
    this.reverseMap = new Map()
    if (compiled.mapping) {
      let colorIdx = 0
      for (const m of compiled.mapping) {
        if (!this.map.has(m.source)) {
          this.map.set(m.source, {
            color: uniqueColors[colorIdx % uniqueColors.length],
            colorIdx: colorIdx % uniqueColors.length,
            ranges: [{ start: m.start, end: m.end }],
          })
        } else {
          this.map.get(m.source)?.ranges.push({ start: m.start, end: m.end })
        }
        for (let i = m.start; i <= m.end; i++) {
          this.reverseMap.set(i, m.source)
        }
        colorIdx++
      }
    }

    this.sourceDecorations = this.getSourceBlockDecorations(compiled)
    this.assemblyDecorations = this.getAssemblyBlockDecorations()
  }

  highlightFromSource(lineNumber: number) {
    if (this.map.has(lineNumber)) {
      this.highlightedSource = new monaco.Range(lineNumber, 1, lineNumber, 1)
      if (this.assembly.code.length > 0) {
        const asmRanges = this.map.get(lineNumber)!.ranges
        const ranges = Array<monaco.Range>()
        for (const r of asmRanges) {
          ranges.push(new monaco.Range(r.start, 1, r.end, 1))
        }
        this.highlightedAssembly = ranges
      }
    } else {
      this.highlightedSource = undefined
      this.highlightedAssembly = undefined
    }
  }

  highlightFromAssembly(lineNumber: number) {
    const sourceLine = this.reverseMap.get(lineNumber)
    if (sourceLine) {
      this.highlightedSource = new monaco.Range(sourceLine, 1, sourceLine, 1)
      const asmRanges = this.map.get(sourceLine)?.ranges || []
      const ranges = Array<monaco.Range>()
      for (const r of asmRanges) {
        ranges.push(new monaco.Range(r.start, 1, r.end, 1))
      }
      this.highlightedAssembly = ranges
    } else {
      this.highlightedSource = undefined
      this.highlightedAssembly = undefined
    }
  }

  getSourceBlockDecorations(compiled: CompilationResult): monaco.editor.IModelDeltaDecoration[] {
    const decs = new Array<monaco.editor.IModelDeltaDecoration>()
    for (const [lineNumber, map] of this.map) {
      decs.push({
        range: new monaco.Range(lineNumber, 1, lineNumber, 1),
        options: {
          isWholeLine: true,
          className: `block-color-${map.colorIdx + 1}`,
        },
      })
    }

    if (compiled.inliningAnalysis) {
      for (const fc of compiled.inliningAnalysis) {
        const [line, column] = [fc.location.l, fc.location.c]
        const decoration = {
          range: new monaco.Range(line, column, line, column + fc.name.length),
          options: {
            hoverMessage: [
              { value: `\`${fc.name}\` ${fc.canInline ? 'can' : 'cannot'} be inlined` },
            ],
            inlineClassName: fc.canInline
              ? 'inline-hover-can-inline'
              : 'inline-hover-cannot-inline',
          },
        }
        if (fc.canInline) {
          decoration.options.hoverMessage.push({ value: `cost: ${fc.cost}` })
        } else {
          decoration.options.hoverMessage.push({ value: fc.reason })
        }
        decs.push(decoration)
      }
    }

    if (compiled.inlinedCalls) {
      for (const ic of compiled.inlinedCalls) {
        const [line, column] = [ic.location.l, ic.location.c]
        decs.push({
          range: new monaco.Range(line, column, line, column + ic.length),
          options: {
            hoverMessage: [{ value: `inlining call to \`${ic.name}\`` }],
            className: 'inlinedcall',
          },
        })
      }
    }

    if (compiled.heapEscapes) {
      for (const he of compiled.heapEscapes) {
        const [line, column] = [he.location.l, he.location.c]
        decs.push({
          range: new monaco.Range(line, column, line, column + he.name.length),
          options: {
            hoverMessage: { value: `\`${he.name}\` escapes to heap` },
            inlineClassName: 'inline-hover-escape',
          },
        })
      }
    }

    return decs
  }

  getAssemblyBlockDecorations(): monaco.editor.IModelDeltaDecoration[] {
    const decs = new Array<monaco.editor.IModelDeltaDecoration>()
    for (const [, map] of this.map) {
      for (const range of map.ranges) {
        decs.push({
          range: new monaco.Range(range.start, 1, range.end, 1),
          options: {
            isWholeLine: true,
            className: `block-color-${map.colorIdx + 1}`,
          },
        })
      }
    }
    return decs
  }
}

function splitAssembly(source: string): Assembly {
  const lines = source.split('\n')
  const code = new Array(lines.length)
  const addresses = new Array(lines.length)
  for (let i = 0; i < lines.length; i++) {
    if (lines[i].startsWith('0x')) {
      ;[addresses[i], code[i]] = lines[i].split('\t')
    } else {
      ;[addresses[i], code[i]] = ['', lines[i]]
    }
  }
  return {
    code: code.join('\n'),
    addresses: addresses,
  }
}

const uniqueColors = [
  '#e6194B',
  '#3cb44b',
  '#ffe119',
  '#4363d8',
  '#f58231',
  '#911eb4',
  '#42d4f4',
  '#f032e6',
  '#bfef45',
  '#fabed4',
  '#469990',
  '#dcbeff',
  '#9A6324',
  '#fffac8',
  '#800000',
  '#aaffc3',
  '#808000',
  '#ffd8b1',
  '#000075',
  '#a9a9a9',
  '#ffffff',
  '#000000',
]
