import EditorJS from '@editorjs/editorjs'
import Header from '@editorjs/header'
import List from '@editorjs/list'

window.isFirstLoad = window.isFirstLoad ?? true;

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

function resizeAndRun(site, el) {
  el.style.height = 'auto';

  const newHeight = el.scrollHeight;
  const lineHeight = parseFloat(getComputedStyle(el).lineHeight);
  const isMultiLine = newHeight > lineHeight * 1.5;

  if (isMultiLine) {
    el.removeAttribute('rows');
  } else {
    el.setAttribute('rows', '1');
  }

  el.style.height = newHeight + 'px';

  if (!site) return;

  const stored = localStorage.getItem(`site:${site}`);
  const data = stored ? JSON.parse(stored) : {};

  // Use localStorage values on first load
  // Use element value on subsequent changes
  if (window.isFirstLoad) {
    if (el.id === 'editor_title' && data.title) {
      el.value = data.title;
    } else if (el.id === 'editor_description' && data.description) {
      el.value = data.description;
    }
  } else {
    if (el.id === 'editor_title') {
      data.title = el.value || "{{ site.SiteTitle }}";
    } else if (el.id === 'editor_description') {
      data.description = el.value;
    }
    localStorage.setItem(`site:${site}`, JSON.stringify(data));
  }

  // resize again after possible content change
  el.style.height = 'auto';
  el.style.height = el.scrollHeight + 'px';
}
window.resizeAndRun = resizeAndRun;
