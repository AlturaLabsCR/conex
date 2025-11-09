package handlers

import (
	"net/http"

	"app/internal/db"
	"app/templates"
)

func (h *Handler) Site(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	siteSlug := r.PathValue("site")

	queries := db.New(h.DB())

	site, err := queries.GetPublishedSiteWithMetricsBySlug(ctx, siteSlug)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		templates.NotFound(h.Translator(r)).Render(ctx, w)
		return
	}

	header := templates.SiteHeader(site)
	content := templates.Site(site)

	templates.Base(h.Translator(r), header, content).Render(ctx, w)
}
