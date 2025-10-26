package handlers

import (
	"net/http"

	"app/internal/db"
	"app/templates"
)

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	queries := db.New(h.DB())

	sites, err := queries.GetValidSitesWithMetricsFromMostTotalVisits(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	header := templates.HomeHeader(h.Translator(r))
	content := templates.CardsGrid(h.Translator(r), sites)

	templates.Base(h.Translator(r), header, content).Render(ctx, w)
}
