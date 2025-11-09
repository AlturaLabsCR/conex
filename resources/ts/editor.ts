import EditorJS from '@editorjs/editorjs'
import Header from '@editorjs/header'
import List from '@editorjs/list'

export function initEditor(site) {
  let siteData = null

  if (site) {
    const stored = localStorage.getItem(`site:${site}`)

    if (stored) {
      try {
        siteData = JSON.parse(stored)
      } catch (e) {
        console.warn("Invalid JSON for site:", site, e)
      }
    }
  }

  const titleEl = document.getElementById('editor_title')
  const descEl = document.getElementById('editor_description')

  const defaultTitle = siteData?.title || ''
  const defaultDesc = siteData?.description || ''

  if (titleEl) {
    titleEl.value = defaultTitle
    titleEl.dispatchEvent(new Event('input'))
  }

  if (descEl) {
    descEl.value = defaultDesc
    descEl.dispatchEvent(new Event('input'))
  }

  const editor = new EditorJS({
    holder: 'editorjs',

    tools: {
      header: Header,
      list: List
    },

    data: siteData?.content || {
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

      if (site) {
        const updated = {
          title: titleEl?.value || '',
          description: descEl?.value || '',
          content: output
        }
        localStorage.setItem(`site:${site}`, JSON.stringify(updated))
      }
    }
  })
}

window.initEditor = initEditor;
