import * as monaco from 'monaco-editor/esm/vs/editor/editor.api'
import { ref, type Ref } from 'vue'

import type { CompilationResult } from '@/services/api'

import './sourcemap.scss'

export interface Mapping {
  color: string
  colorIdx: number
  ranges: {
    start: number
    end: number
  }[]
}

export class SourceMap {
  map: Map<number, Mapping>
  reverseMap: Map<number, number>
  highlightedRange: Ref<number>
  inliningAnalysis: CompilationResult['inliningAnalysis']
  heapEscapes: CompilationResult['heapEscapes']

  constructor(compiled: CompilationResult) {
    let colorIdx = 0
    this.map = new Map()
    this.reverseMap = new Map()
    this.highlightedRange = ref(-1)
    if (compiled.mapping) {
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

    this.inliningAnalysis = compiled.inliningAnalysis || []
    this.heapEscapes = compiled.heapEscapes || []
  }

  highlightFromSource(lineNumber: number) {
    if (this.map.has(lineNumber)) {
      if (this.highlightedRange.value != lineNumber) {
        this.highlightedRange.value = lineNumber
      }
      return
    }
    this.highlightedRange.value = -1
  }

  highlightFromAssembly(lineNumber: number) {
    const sourceLine = this.reverseMap.get(lineNumber)
    if (sourceLine) {
      if (this.highlightedRange.value != sourceLine) {
        this.highlightedRange.value = sourceLine
      }
      return
    }
    this.highlightedRange.value = -1
  }

  sourceBlockDecorations(): monaco.editor.IModelDeltaDecoration[] {
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
    for (const fc of this.inliningAnalysis) {
      const [line, column] = [fc.location.l, fc.location.c]
      const decoration = {
        range: new monaco.Range(line, column, line, column + fc.name.length),
        options: {
          hoverMessage: [
            { value: `# \`${fc.name}\` ${fc.canInline ? 'can' : 'cannot'} be inlined` },
          ],
          inlineClassName: fc.canInline ? 'inline-hover-can-inline' : 'inline-hover-cannot-inline',
        },
      }
      if (fc.canInline) {
        decoration.options.hoverMessage.push({ value: `cost: ${fc.cost}` })
      } else {
        decoration.options.hoverMessage.push({ value: fc.reason })
      }
      decs.push(decoration)
    }
    for (const he of this.heapEscapes) {
      const [line, column] = [he.location.l, he.location.c]
      decs.push({
        range: new monaco.Range(line, column, line, column + he.name.length),
        options: {
          hoverMessage: { value: `\`${he.name}\` escapes to heap` },
          inlineClassName: 'inline-hover-escape',
        },
      })
    }
    return decs
  }

  assemblyBlockDecorations(): monaco.editor.IModelDeltaDecoration[] {
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
