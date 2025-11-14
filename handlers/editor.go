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

	session, ok := ctx.Value(ctxSessionKey).(db.Session)
	if !ok {
		h.Log().Error("error retrieving session from ctx")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s := r.PathValue("site")

	queries := db.New(h.DB())

	site, err := queries.GetSiteWithMetrics(ctx, s)
	if err != nil {
		h.Log().Error("error loading sites", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if site.SiteUser != session.SessionUser {
		h.Log().Debug("tried to load a site without ownership", "user_id", session.SessionUser, "site_slug", site.SiteSlug)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	tr := h.Translator(r)

	header := templates.EditorHeader(tr, site, "")
	content := templates.Editor(tr, site)

	templates.Base(tr, header, content, true).Render(ctx, w)
}

func (h *Handler) Publish(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tr := h.Translator(r)

	session, ok := ctx.Value(ctxSessionKey).(db.Session)
	if !ok {
		h.Log().Error("error retrieving session from ctx")
		templates.Notice(
			templates.PublishNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}

	type PublishData struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Slug        string `json:"slug"`
		Content     string `json:"content"`
	}

	var data PublishData

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		h.Log().Error("invalid publish request", "data", data)
		templates.Notice(
			templates.PublishNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}
	defer r.Body.Close()

	if data.Title == "" {
		h.Log().Error("title is empty")
		templates.Notice(
			templates.PublishNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("editor_empty_title"),
		).Render(ctx, w)
		return
	}

	queries := db.New(h.DB())

	site, err := queries.GetSiteBySlug(ctx, data.Slug)
	if err != nil {
		h.Log().Error("error querying site", "error", err)
		templates.Notice(
			templates.PublishNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}

	if site.SiteUser != session.SessionUser {
		w.WriteHeader(http.StatusUnauthorized)
		h.Log().Error("user does not own site", "site_user", session.SessionUser)
		return
	}

	queries.UpdateSite(ctx, db.UpdateSiteParams{
		SiteID:            site.SiteID,
		SiteTitle:         data.Title,
		SiteDescription:   data.Description,
		SiteTagsJson:      site.SiteTagsJson,
		SiteHtmlPublished: data.Content,
		SiteModifiedUnix:  time.Now().Unix(),
		SitePublished:     1,
		SiteDeleted:       0,
	})

	h.Log().Debug("updated site", "site_id", site.SiteID, "site_html_published", data.Content)

	templates.Notice(
		templates.PublishNoticeID,
		templates.NoticeInfo,
		tr("success"),
		tr("dashboard_published_site"),
	).Render(ctx, w)
}
