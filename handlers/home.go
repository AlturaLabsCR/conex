package handlers

import (
	"net/http"

	"app/internal/db"
	"app/templates"
)

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	queries := db.New(h.DB())

	sites, err := queries.GetHomePageSitesWithMetricsFromMostTotalVisits(ctx)
	if err != nil {
		h.Log().Error("error querying sites with metrics", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tr := h.Translator(r)

	header := templates.HomeHeader(tr)
	content := templates.CardsGrid(tr, sites, false)

	if err := templates.Base(tr, header, content, false).Render(ctx, w); err != nil {
		h.Log().Error("error rendering template", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
