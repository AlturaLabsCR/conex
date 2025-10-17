package handlers

import (
	"net/http"

	"app/templates"
)

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	header := templates.HomeHeader()
	content := templates.Home()

	templates.Base(header, content).Render(ctx, w)
}
