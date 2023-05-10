import mitt from 'mitt'

type Events = {
  formatCode: void,
  compileCode: void,
  shareCode: void,

  editorLayoutRequested: void
  sourceLineHovered: number
  assemblyLineHovered: number

  jumpToAssemblyLine: number
  jumpToSourceLine: number
}

const bus = mitt<Events>()

export default bus
