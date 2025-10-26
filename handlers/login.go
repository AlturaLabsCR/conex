package handlers

import (
	"net/http"

	"app/templates"
)

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	header := templates.LoginHeader(h.Translator(r))
	content := templates.Login(h.Translator(r))

	templates.Base(h.Translator(r), header, content).Render(ctx, w)
}
