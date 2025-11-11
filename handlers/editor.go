package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"app/internal/db"
	"app/templates"
)

func (h *Handler) Editor(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	s := r.PathValue("site")

	queries := db.New(h.DB())

	site, err := queries.GetSiteWithMetrics(ctx, s)
	if err != nil {
		h.Log().Error("error loading sites", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tr := h.Translator(r)

	header := templates.EditorHeader(tr, site, "")
	content := templates.Editor(tr, site)

	templates.Base(tr, header, content).Render(ctx, w)
}

func (h *Handler) Publish(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	type PublishData struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Slug        string `json:"slug"`
		Content     string `json:"content"`
	}

	var data PublishData

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	queries := db.New(h.DB())

	site, _ := queries.GetSiteBySlug(ctx, data.Slug)

	queries.UpdateSite(ctx, db.UpdateSiteParams{
		SiteID:            site.SiteID,
		SiteTitle:         data.Title,
		SiteDescription:   data.Description,
		SiteTagsJson:      "",
		SiteHtmlPublished: data.Content,
		SiteModifiedUnix:  time.Now().Unix(),
		SitePublished:     1,
		SiteDeleted:       0,
	})

	h.Log().Debug("updated site", "site_id", site.SiteID, "site_html_published", data.Content)
}
