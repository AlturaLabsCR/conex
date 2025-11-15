package handlers

import (
	"compress/gzip"
	"net/http"
	"strings"

	"app/internal/db"
	"app/templates"
)

func (h *Handler) Site(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	siteSlug := r.PathValue("site")

	queries := db.New(h.DB())

	site, err := queries.GetPublishedSiteWithMetricsBySlug(ctx, siteSlug)
	if err != nil || site.SitePublished != 1 {
		h.Log().Debug("cannot find published site with metrics", "siteSlug", siteSlug, "sitePublished", site.SitePublished)
		if err := templates.NotFound(h.Translator(r)).Render(ctx, w); err != nil {
			h.Log().Error("error rendering template", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}

	isOwner := false
	if s, err := h.verifyClient(
		w,
		r,
		false,
	); s.SessionUser == site.SiteUser && err == nil {
		h.Log().Debug("is owner")
		isOwner = true
	}

	tr := h.Translator(r)

	header := templates.SiteHeader(tr, site, "", isOwner)
	content := templates.Site(site)

	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		gz := gzip.NewWriter(w)
		defer gz.Close()
		w.Header().Add("Content-Type", "text/html")
		w.Header().Add("Content-Encoding", "gzip")
		templates.Base(h.Translator(r), header, content, false).Render(ctx, gz)
	} else {
		templates.Base(h.Translator(r), header, content, false).Render(ctx, w)
	}
}
