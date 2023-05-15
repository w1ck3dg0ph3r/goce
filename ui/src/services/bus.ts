import mitt from 'mitt'

type Events = {
  formatCode: void
  compileCode: void
  shareCode: void

  sourceLineHovered: number
  assemblyLineHovered: number

  revealSourceLine: number
  jumpToSourceLine: {
    line: number
    column?: number
  }
  revealAssemblyLine: number
}

const bus = mitt<Events>()

export default bus
