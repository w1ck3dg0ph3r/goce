import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker'

self.MonacoEnvironment = {
  getWorker(id: string, label: string) {
    return new editorWorker()
  },
}
export {}
