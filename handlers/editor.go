package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"app/database"
	"app/internal/db"
	"app/templates"
)

type SyncResponse struct {
	ShouldPatch bool            `json:"shouldPatch"`
	SiteData    json.RawMessage `json:"siteData,omitempty"`
}

type SyncRequest struct {
	LocalData database.SiteData `json:"localData"`
}

func (h *Handler) Editor(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	session, ok := ctx.Value(ctxSessionKey).(db.Session)
	if !ok {
		h.Log().Debug("error retrieving session from ctx")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s := r.PathValue("site")

	queries := db.New(h.DB())

	site, err := queries.GetSiteWithMetrics(ctx, s)
	if err != nil {
		h.Log().Error("error loading site", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if site.SiteUser != session.SessionUser {
		h.Log().Debug(
			"tried to load a site without ownership",
			"user_id",
			session.SessionUser,
			"site_slug",
			site.SiteSlug,
		)
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
		h.Log().Debug("error retrieving session from ctx")
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
		h.Log().Debug("invalid publish request", "data", data)
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
		h.Log().Debug("title is empty")
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
		h.Log().Debug("user does not own site", "site_user", session.SessionUser)
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

	if err := templates.Notice(
		templates.PublishNoticeID,
		templates.NoticeInfo,
		tr("success"),
		tr("dashboard_published_site"),
	).Render(ctx, w); err != nil {
		h.Log().Error("error rendering template", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) EditorSync(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	slug := r.PathValue("site")

	if slug == "" {
		h.Log().Error("error invalid slug", "slug", slug)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h.Log().Debug("site slug is valid")

	session, ok := ctx.Value(ctxSessionKey).(db.Session)
	if !ok {
		h.Log().Debug("error retrieving session from ctx")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	h.Log().Debug("session id valid")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.Log().Error("error reading patch sync body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	h.Log().Debug("body is readable")

	var req SyncRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.Log().Error("error invalid patch sync body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h.Log().Debug("body is valid")

	tx, err := h.DB().Begin()
	if err != nil {
		h.Log().Error("error starting tx", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()
	h.Log().Debug("started tx")

	queries := db.New(h.DB()).WithTx(tx)

	site, err := queries.GetSiteWithMetrics(ctx, slug)
	if err != nil {
		h.Log().Error("error querying site with metrics", "error", err, "slug", slug)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.Log().Debug("site exists")

	if site.SiteUser != session.SessionUser {
		h.Log().Error("error is not site owner")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	h.Log().Debug("session user is site owner")

	serverData, err := queries.GetSyncData(ctx, site.SiteID)
	if err != nil {
		h.Log().Debug("server data does not exist")
		h.Log().Debug("running patch sync on site", "slug", slug)

		b, err := json.Marshal(req)
		if err != nil {
			h.Log().Debug("error marshalling req")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if _, err := queries.InsertSyncData(ctx, db.InsertSyncDataParams{
			SiteSyncID:             site.SiteID,
			SiteSyncDataStaging:    string(b),
			SiteSyncLastUpdateUnix: time.Now().Unix(),
		}); err != nil {
			h.Log().Debug("error inserting client data")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			h.Log().Error("commit failed", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		h.Log().Debug("inserted sync data, returning patch false as client is up-to-date")

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(SyncResponse{
			ShouldPatch: false,
		}); err != nil {
			h.Log().Error("error encoding json", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}

	h.Log().Debug("server data exists")

	var resp SyncResponse
	if req.LocalData.LastUpdated > serverData.SiteSyncLastUpdateUnix {
		h.Log().Debug("client is newer, update server")
		b, _ := json.Marshal(req)
		if err := queries.UpdateSyncData(ctx, db.UpdateSyncDataParams{
			SiteSyncID:             site.SiteID,
			SiteSyncDataStaging:    string(b),
			SiteSyncLastUpdateUnix: req.LocalData.LastUpdated,
		}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		h.Log().Debug("server updated")
		resp.ShouldPatch = false
	} else {
		h.Log().Debug("server is newer, siteData in response")
		resp.ShouldPatch = true
		resp.SiteData = json.RawMessage(serverData.SiteSyncDataStaging)
	}

	if err := tx.Commit(); err != nil {
		h.Log().Error("commit failed", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.Log().Debug("ended tx, responding")

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.Log().Error("json encode failed", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
