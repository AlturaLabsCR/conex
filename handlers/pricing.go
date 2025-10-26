package handlers

import (
	"net/http"

	"app/templates"
)

func (h *Handler) Pricing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	header := templates.PricingHeader(h.Translator((r)))
	content := templates.Pricing(h.Translator((r)))

	templates.Base(h.Translator(r), header, content).Render(ctx, w)
}
