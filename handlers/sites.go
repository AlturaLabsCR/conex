package handlers

import (
	"compress/gzip"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"app/config"
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

	bannerURL := ""

	banner, err := queries.GetBanner(ctx, site.SiteID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			h.Log().Error("error loading site", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		object, err := queries.GetObjectByID(ctx, banner.BannerObject)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				h.Log().Error("error loading site", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		bannerURL = config.S3PublicURL + "/" + object.ObjectKey
	}

	tr := h.Translator(r)

	header := templates.SiteHeader(tr, site, bannerURL, isOwner)
	content := templates.Site(site)

	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		gz := gzip.NewWriter(w)
		defer gz.Close()
		w.Header().Add("Content-Type", "text/html")
		w.Header().Add("Content-Encoding", "gzip")
		templates.Base(tr, header, content, false).Render(ctx, gz)
	} else {
		templates.Base(tr, header, content, false).Render(ctx, w)
	}

	if err := queries.NewVisit(ctx, site.SiteID); err != nil {
		h.Log().Error("error incrementing visit", "error", err)
		return
	}
}
