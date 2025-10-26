package handlers

import (
	"net/http"

	"app/templates"

	"github.com/a-h/templ"
)

func (h *Handler) Editor(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var header templ.Component
	var content templ.Component

	if r.PathValue("site") == "" {
		header = templates.EditorHeader(h.Translator((r)))
		content = templates.Editor()
	} else {
		// TODO: fill with site data
	}

	templates.Base(h.Translator(r), header, content).Render(ctx, w)
}
