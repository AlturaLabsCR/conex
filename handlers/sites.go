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

	tr := h.Translator(r)

	header := templates.SiteHeader(tr, site, "", true)
	content := templates.Site(site)

	// gz := gzip.NewWriter(w)
	// defer gz.Close()
	// w.Header().Add("Content-Type", "text/html")
	// w.Header().Add("Content-Encoding", "gzip")

	templates.Base(h.Translator(r), header, content).Render(ctx, w)
}
