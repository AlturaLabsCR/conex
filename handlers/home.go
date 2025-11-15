package handlers

import (
	"net/http"

	"app/internal/db"
	"app/templates"
)

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	queries := db.New(h.DB())

	sites, err := queries.GetPublishedSitesWithMetricsFromMostTotalVisits(ctx)
	if err != nil {
		h.Log().Error("error querying sites with metrics", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	header := templates.HomeHeader(h.Translator(r))
	content := templates.CardsGrid(h.Translator(r), sites, false)

	if err := templates.Base(h.Translator(r), header, content, false).Render(ctx, w); err != nil {
		h.Log().Error("error rendering template", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
