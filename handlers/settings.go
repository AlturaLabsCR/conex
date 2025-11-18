package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"app/database"
	"app/internal/db"
	"app/templates"
)

type updateSettingsRequest struct {
	Slug     string `json:"slug"`
	Tags     string `json:"tags,omitempty"`
	HomePage string `json:"home_page,omitempty"`
}

func (h *Handler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tr := h.Translator(r)

	var req updateSettingsRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Log().Error("failed to decode update settings req", "error", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.Slug == "" {
		templates.Notice(
			templates.UpdateSettingsNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("editor_bad_tags"),
		).Render(ctx, w)
		return
	}

	tags, err := database.ParseTags(req.Tags)
	if err != nil {
		h.Log().Debug("failed to upadate tags", "error", err)
		templates.Notice(
			templates.UpdateSettingsNoticeID,
			templates.NoticeWarn,
			tr("warn"),
			tr("editor_tag_limit"),
		).Render(ctx, w)
		return
	}

	if len(tags) > 3 {
		h.Log().Debug("failed to upadate tags", "error", "more than 3 tags")
		templates.Notice(
			templates.UpdateSettingsNoticeID,
			templates.NoticeWarn,
			tr("warn"),
			tr("editor_tag_limit"),
		).Render(ctx, w)
		return
	}

	if req.HomePage == "" && len(tags) == 0 {
		h.Log().Debug("empty request, not doing anything")
		return
	}

	session, ok := ctx.Value(ctxSessionKey).(db.Session)
	if !ok {
		h.Log().Debug("error retrieving session from ctx")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tx, err := h.DB().Begin(ctx)
	if err != nil {
		h.Log().Error("error starting tx", "error", err)
		templates.Notice(
			templates.UpdateSettingsNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}
	defer tx.Rollback(ctx)

	site, err := h.Queries().GetSiteWithMetrics(ctx, req.Slug)
	if err != nil {
		h.Log().Error("error querying site by slug", "error", err)
		templates.Notice(
			templates.UpdateSettingsNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}

	if site.SiteUser != session.SessionUser {
		h.Log().Error("cannot update tags, not owner of the site")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var showHomePage int64

	switch req.HomePage {
	case "show":
		showHomePage = 1
	case "no":
		showHomePage = 0
	default:
		showHomePage = site.SiteHomePage
	}

	json := database.TagsToJSON(tags)

	if database.TagsToCommaList(json) == database.TagsToCommaList(site.SiteTagsJson) {
		json = site.SiteTagsJson
	}

	h.Log().Debug("updating tags", "tags", tags, "json", json)

	if err := h.Queries().UpdateSiteSettings(ctx, db.UpdateSiteSettingsParams{
		SiteTagsJson:     json,
		SiteModifiedUnix: time.Now().Unix(),
		SiteHomePage:     showHomePage,
		SiteID:           site.SiteID,
	}); err != nil {
		h.Log().Debug("error updating tags", "error", err)
		templates.Notice(
			templates.UpdateSettingsNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}

	if err := templates.ShowInHomeButton(
		tr,
		showHomePage == 1,
		site.SitePublished == 1,
	).Render(ctx, w); err != nil {
		h.Log().Error("error rendering new shown status")
	}

	if err := tx.Commit(ctx); err != nil {
		h.Log().Debug("error commit tx", "error", err)
		templates.Notice(
			templates.UpdateSettingsNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}

	site.SiteTagsJson = json

	if err := templates.EditorTags(tr, site).Render(ctx, w); err != nil {
		h.Log().Error("error rendering template", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	templates.NoticeEmpty(templates.UpdateSettingsNoticeID).Render(ctx, w)
}
