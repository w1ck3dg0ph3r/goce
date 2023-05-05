import mitt from 'mitt'

type Events = {
  editorLayoutRequested: void
  sourceLineHovered: number
  assemblyLineHovered: number

  jumpToAssemblyLine: number
  jumpToSourceLine: number
}

const bus = mitt<Events>()

export default bus
