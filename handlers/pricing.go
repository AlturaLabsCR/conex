package handlers

import (
	"net/http"

	"app/templates"
)

func (h *Handler) Pricing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	header := templates.PricingHeader(h.Translator((r)))
	content := templates.Pricing(h.Translator((r)))

	if err := templates.Base(h.Translator(r), header, content, true).Render(ctx, w); err != nil {
		h.Log().Error("error rendering template", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
