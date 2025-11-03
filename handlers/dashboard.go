package handlers

import (
	"net/http"

	"app/internal/db"
	"app/templates"
)

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	header := templates.DashboardHeader(h.Translator(r))
	content := templates.CardsGrid(h.Translator(r), []db.ValidSitesWithMetric{})

	templates.Base(h.Translator(r), header, content).Render(ctx, w)
}
