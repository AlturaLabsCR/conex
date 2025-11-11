import EditorJS from '@editorjs/editorjs'
import Header from '@editorjs/header'
import List from '@editorjs/list'
import Table from '@editorjs/table'
import edjsHTML from 'editorjs-html'

interface SiteData {
  title?: string;
  description?: string;
  content?: any;
}

let isFirstLoad = false

const edjsParser = edjsHTML({
  table: tableParser,
});

export async function initEditor(site: string) {
  let siteData: SiteData | null = null;

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

  const defaultTitle = siteData?.title || ''
  const defaultDesc = siteData?.description || ''

  const titleEl = document.getElementById('editor_title') as HTMLTextAreaElement | null;
  if (titleEl) {
    titleEl.value = defaultTitle
    titleEl.dispatchEvent(new Event('input'))
  }

  const descEl = document.getElementById('editor_description') as HTMLTextAreaElement | null;
  if (descEl) {
    descEl.value = defaultDesc
    descEl.dispatchEvent(new Event('input'))
  }

  const editor = new EditorJS({
    holder: 'editorjs',

    tools: {
      header: Header,
      list: List,
      table: Table,
    },

    placeholder: "Lorem ipsum dolor sit amet.",

    data: siteData?.content || {
      time: Date.now(),
      blocks: [
        {
          type: "paragraph",
          data: {
            text: ""
          }
        }
      ]
    },

    onReady: async () => {
      const output = await editor.save()
      const outputHTML = edjsParser.parse(output)
      const htmlEl = document.getElementById('editor_html') as HTMLTextAreaElement | null
      if (htmlEl) {
        htmlEl.value = Array.isArray(outputHTML) ? outputHTML.join('') : String(outputHTML)
      }
    },

    onChange: async () => {
      const output = await editor.save()

      if (site) {
        const updated = {
          title: titleEl?.value || '',
          description: descEl?.value || '',
          content: output
        }

        localStorage.setItem(`site:${site}`, JSON.stringify(updated))

        const outputHTML = edjsParser.parse(output)
        const htmlEl = document.getElementById('editor_html') as HTMLTextAreaElement | null;
        if (htmlEl) {
          htmlEl.value = Array.isArray(outputHTML) ? outputHTML.join('') : String(outputHTML)
        }
      }
    }
  });
}
(window as any).initEditor = initEditor;

function resizeAndRun(site: string, el: HTMLTextAreaElement) {
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
  if (isFirstLoad) {
    if (el.id === 'editor_title' && data.title) {
      el.value = data.title;
    } else if (el.id === 'editor_description' && data.description) {
      el.value = data.description;
    }
  } else {
    if (el.id === 'editor_title') {
      data.title = el.value;
    } else if (el.id === 'editor_description') {
      data.description = el.value;
    }
    localStorage.setItem(`site:${site}`, JSON.stringify(data));
  }

  // resize again after possible content change
  el.style.height = 'auto';
  el.style.height = el.scrollHeight + 'px';
}
(window as any).resizeAndRun = resizeAndRun;

export function getEditorHtml(): string {
  const htmlEl = document.getElementById('editor_html') as HTMLTextAreaElement | null;
  return htmlEl?.value || '';
}
(window as any).getEditorHtml = getEditorHtml;

function tableParser(block) {
  const { withHeadings, content } = block.data;

  if (!Array.isArray(content) || content.length === 0) {
    return "";
  }

  let html = "<table>";

  // if the first row should be used as a header
  if (withHeadings) {
    const headers = content[0];
    html += "<thead><tr>";
    headers.forEach(cell => {
      html += `<th>${cell}</th>`;
    });
    html += "</tr></thead>";

    // the rest of the rows go in the body
    if (content.length > 1) {
      html += "<tbody>";
      for (let i = 1; i < content.length; i++) {
        html += "<tr>";
        content[i].forEach(cell => {
          html += `<td>${cell}</td>`;
        });
        html += "</tr>";
      }
      html += "</tbody>";
    }
  } else {
    // no headings, everything goes in <tbody>
    html += "<tbody>";
    content.forEach(row => {
      html += "<tr>";
      row.forEach(cell => {
        html += `<td>${cell}</td>`;
      });
      html += "</tr>";
    });
    html += "</tbody>";
  }

  html += "</table>";
  return html;
}
