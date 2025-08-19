import * as monaco from 'monaco-editor/esm/vs/editor/editor.api'

import registers386 from './registers_386'
import registersAMD64 from './registers_amd64'
import registersARM from './registers_arm'
import registersARM64 from './registers_arm64'
import registersPPC64 from './registers_ppc64'

const languageId = 'plan9asm'

const language: monaco.languages.IMonarchLanguage = {
  defaultToken: '',
  registers: Array<string>().concat(
    registers386,
    registersAMD64,
    registersARM,
    registersARM64,
    registersPPC64
  ),
  tokenizer: {
    root: [
      [/^([^\s]+)\s*/, 'keyword'],
      [
        /[A-Za-z]+[\w_.·:*+=]+/,
        {
          cases: {
            '@registers': 'string',
            '@default': 'identifier',
          },
        },
      ],
      [/[ \t,()\r\n]+/, 'white'],
      [/[,()]/, ''],
      [/(\$?[-0-9xA-Fa-f]+)/, 'number'],
    ],
  },
}

monaco.languages.register({ id: languageId })

monaco.languages.setMonarchTokensProvider(languageId, language)

monaco.languages.registerFoldingRangeProvider(languageId, {
  provideFoldingRanges(model) {
    const regions = new Array<monaco.languages.FoldingRange>()
    let region: monaco.languages.FoldingRange | null = null
    for (let i = 0; i < model.getLineCount(); i++) {
      const line = model.getLineContent(i + 1)
      if (line.startsWith('TEXT')) {
        if (!region) {
          region = {
            start: i + 1,
            end: 0,
            kind: monaco.languages.FoldingRangeKind.Region,
          }
        } else {
          region.end = i
          regions.push(region)
          region = {
            start: i + 1,
            end: 0,
            kind: monaco.languages.FoldingRangeKind.Region,
          }
        }
      }
    }
    if (region) {
      region.end = model.getLineCount()
      regions.push(region)
    }
    return regions
  },
})
