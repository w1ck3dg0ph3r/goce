import mitt from 'mitt'

type Events = {
  editorLayoutRequested: void
}

const bus = mitt<Events>()

export default bus
