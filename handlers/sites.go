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

	site, err := queries.GetValidSiteBySlug(ctx, siteSlug)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		// TODO: 404 page
		return
	}

	header := templates.SiteHeader(site)
	content := templates.Site(site)

	templates.Base(header, content).Render(ctx, w)
}
