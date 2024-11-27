
import { Controlled as CodeMirror, ICodeMirror } from 'react-codemirror2'
import { Editor, EditorConfiguration } from 'codemirror'


import 'codemirror/lib/codemirror.css'
import 'codemirror/theme/base16-light.css'
import 'codemirror/mode/toml/toml'
import 'codemirror/mode/javascript/javascript'
import 'codemirror/addon/selection/active-line'
import 'codemirror/addon/scroll/scrollpastend'
import { useEffect, useState } from 'react'

const setEditorOverlay = (editor: Editor) => {
  const query = /\{\{[^}]*}}/g
  editor.addOverlay({
    token: function (stream: any) {
      query.lastIndex = stream.pos
      var match = query.exec(stream.string)
      if (match && match.index === stream.pos) {
        stream.pos += match[0].length || 1

        return 'notelink'
      } else if (match) {
        stream.pos = match.index
      } else {
        stream.skipToEnd()
      }
    },
  })
}

export interface DCAEditorProps extends ICodeMirror {
  value: string,
  setValue?: (value: string) => void,
  editorOptions?: EditorConfiguration,
}

export interface DCAEditorConfiguration extends EditorConfiguration { }

const defaultOptions: DCAEditorConfiguration = {
  mode: "toml",
  lineNumbers: true,
  lineWrapping: true,
  dragDrop: false,
  readOnly: true,
  cursorBlinkRate: 0,
}

export default function DCAEditor({ value, setValue, editorOptions, editorDidMount }: DCAEditorProps) {
  const mergedOptions = { ...defaultOptions, ...editorOptions }
  const [options, setOptions] = useState(mergedOptions)

  useEffect(() => {
    const mergedOptions = { ...defaultOptions, ...editorOptions }
    if (mergedOptions.readOnly) {
      mergedOptions.cursorBlinkRate = 0
    } else {
      mergedOptions.cursorBlinkRate = 500
    }
    setOptions(mergedOptions)

  }, [editorOptions])

  return <CodeMirror
    className='editor'
    onBeforeChange={(editor, data, value) => {
      setValue && setValue(value)
    }}
    editorDidMount={(editor, value, cb) => {
      setTimeout(() => {
        editor.focus()
      }, 0)
      editor.setCursor(0)
      setEditorOverlay(editor)
      editorDidMount && editorDidMount(editor, value, cb)
    }}
    onChange={(editor, data, value) => {
      if (!value) {
        editor.focus()
      }
    }}
    onPaste={(editor, event: any) => {
      // Get around pasting issue
      // https://github.com/scniro/react-codemirror2/issues/77
      if (!event.clipboardData || !event.clipboardData.items || !event.clipboardData.items[0])
        return
      event.clipboardData.items[0].getAsString((pasted: any) => {
        if (editor.getSelection() !== pasted) return
        const { anchor, head } = editor.listSelections()[0]
        editor.setCursor({
          line: Math.max(anchor.line, head.line),
          ch: Math.max(anchor.ch, head.ch),
        })
      })
    }}

    value={value}
    options={options}
  />

}