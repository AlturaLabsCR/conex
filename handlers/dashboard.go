package handlers

import (
	"net/http"
	"time"

	"app/config"
	"app/internal/db"
	"app/templates"
)

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	session, ok := ctx.Value(ctxSessionKey).(db.Session)
	if !ok {
		h.Log().Error("error retrieving session from ctx")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	queries := db.New(h.DB())

	sites, err := queries.GetSitesWithMetricsByUserID(ctx, session.SessionUser)
	if err != nil {
		h.Log().Error("error loading sites", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.Log().Debug("loaded user sites", "count", len(sites))

	header := templates.DashboardHeader(h.Translator(r))
	content := templates.Dashboard(h.Translator(r), sites)

	templates.Base(h.Translator(r), header, content).Render(ctx, w)
}

func (h *Handler) NewSite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// TODO:
	// - Check this site can be created given the user's permission
	// - Check site endpoint is not taken or used by the app's functionality
	// - Check site endpoint is alphanumeric, normalized and spaces as dashes
	// - Check site belongs to the user, if existent

	session, ok := ctx.Value(ctxSessionKey).(db.Session)
	if !ok {
		h.Log().Error("error retrieving session from ctx")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	name := r.FormValue("name")
	endpoint := r.FormValue("endpoint")

	tx, err := h.DB().Begin()
	if err != nil {
		h.Log().Error("error starting tx", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	queries := db.New(h.DB()).WithTx(tx)

	siteID, _ := queries.InsertSite(ctx, db.InsertSiteParams{
		SiteUser:          session.SessionUser,
		SiteSlug:          endpoint,
		SiteTitle:         name,
		SiteTagsJson:      "",
		SiteDescription:   "",
		SiteHtmlPublished: "",
		SiteCreatedUnix:   time.Now().Unix(),
		SiteModifiedUnix:  time.Now().Unix(),
		SitePublished:     0,
		SiteDeleted:       0,
	})

	queries.InsertMetric(ctx, db.InsertMetricParams{
		MetricSite:        siteID,
		MetricVisitsTotal: 0,
	})

	tx.Commit()

	templates.Redirect(config.EditorPath+endpoint).Render(ctx, w)
}
