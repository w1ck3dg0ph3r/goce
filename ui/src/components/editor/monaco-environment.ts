import 'monaco-editor/esm/vs/editor/editor.all'

self.MonacoEnvironment = {
  getWorker: function (_, label) {
    const getWorkerModule = (moduleUrl: string, label: string): Worker => {
      const workerUrl = self.MonacoEnvironment?.getWorkerUrl?.(moduleUrl, label)
      if (!workerUrl) throw Error('cannot get getWorkerUrl')
      return new Worker(workerUrl!, {
        name: label,
        type: 'module'
      })
    }

    return getWorkerModule('/monaco-editor/esm/vs/editor/editor.worker?worker', label)
  }
}

export {}