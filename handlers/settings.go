package handlers

import (
	"net/http"

	"app/database"
	"app/internal/db"
	"app/templates"
)

func (h *Handler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tr := h.Translator(r)

	tagsStr := r.FormValue(templates.EditorTagName)
	tags := database.ParseTags(tagsStr)
	if len(tags) > 3 {
		templates.Notice(
			templates.UpdateSettingsNoticeID,
			templates.NoticeWarn,
			tr("warn"),
			tr("editor_tag_limit"),
		).Render(ctx, w)
		return
	}

	slugStr := r.FormValue(templates.EditorSlugName)
	if tagsStr == "" {
		templates.Notice(
			templates.UpdateSettingsNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("editor_bad_tags"),
		).Render(ctx, w)
		return
	}

	session, ok := ctx.Value(ctxSessionKey).(db.Session)
	if !ok {
		h.Log().Error("error retrieving session from ctx")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tx, err := h.DB().Begin()
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
	defer tx.Rollback()

	queries := db.New(h.DB()).WithTx(tx)

	site, err := queries.GetSiteWithMetrics(ctx, slugStr)
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

	json := database.TagsToJSON(tags)

	h.Log().Debug("updating tags", "tags", tags, "json", json)

	if err := queries.UpdateTags(ctx, db.UpdateTagsParams{
		SiteTagsJson: json,
		SiteID:       site.SiteID,
	}); err != nil {
		templates.Notice(
			templates.UpdateSettingsNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}

	if err := tx.Commit(); err != nil {
		templates.Notice(
			templates.UpdateSettingsNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}

	site.SiteTagsJson = json

	templates.EditorTags(tr, site).Render(ctx, w)
}
