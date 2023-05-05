import mitt from 'mitt'

type Events = {
  editorLayoutRequested: void
  sourceLineHovered: number
  assemblyLineHovered: number
}

const bus = mitt<Events>()

export default bus
