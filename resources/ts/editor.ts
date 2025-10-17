import EditorJS from '@editorjs/editorjs'
// import Header from '@editorjs/header'
// import List from '@editorjs/list'

export function initEditor() {
  const editor = new EditorJS({
    holder: 'editorjs',

    // tools: {
    //   header: Header,
    //   list: List
    // },

    data: {
      time: Date.now(),
      blocks: [
        {
          type: "paragraph",
          data: {
            text: "Hello, this is Editor.js!"
          }
        }
      ]
    },

    onReady: () => console.log("Editor.js ready"),

    onChange: async () => {
      const output = await editor.save()
      console.log("Data:", output)
    }

  })
}

document.addEventListener('DOMContentLoaded', initEditor)
