package handlers

import (
	"net/http"

	"app/config"
	"app/templates"
)

func (h *Handler) Pricing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tr := h.Translator(r)

	header := templates.PricingHeader(tr)
	content := templates.Pricing(tr, config.PayPalClientID)

	// order, err := CreateOrder("USD", "19.99")
	// if err != nil {
	// 	h.Log().Debug("access token error", "error", err)
	// } else {
	// 	h.Log().Debug("created order", "order", order)
	// }

	if err := templates.Base(tr, header, content, true).Render(ctx, w); err != nil {
		h.Log().Error("error rendering template", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
