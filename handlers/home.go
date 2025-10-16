package handlers

import (
	"net/http"

	"app/templates"
)

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	content := templates.Home()

	templates.Base(content).Render(ctx, w)
}
