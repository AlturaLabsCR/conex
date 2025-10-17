package handlers

import (
	"net/http"

	"app/templates"
)

func (h *Handler) Editor(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	header := templates.EditorHeader()
	content := templates.Editor()

	templates.Base(header, content).Render(ctx, w)
}
