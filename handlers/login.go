package handlers

import (
	"net/http"

	"app/templates"
)

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	header := templates.LoginHeader()
	content := templates.Login()

	templates.Base(header, content).Render(ctx, w)
}
