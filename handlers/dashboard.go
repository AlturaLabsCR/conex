package handlers

import (
	"errors"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"app/config"
	"app/internal/db"
	"app/templates"
)

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	h.Log().Debug("hit endpoint", "pattern", r.Pattern)

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

	header := templates.DashboardHeader(h.Translator(r))
	content := templates.Dashboard(h.Translator(r), sites)

	if err := templates.Base(h.Translator(r), header, content, true).Render(ctx, w); err != nil {
		h.Log().Error("error rendering template", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) NewSite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tr := h.Translator(r)

	session, ok := ctx.Value(ctxSessionKey).(db.Session)
	if !ok {
		h.Log().Error("error retrieving session from ctx")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	name := r.FormValue("name")
	endpointRaw := r.FormValue("endpoint")

	endpoint, err := parseEndpoint(endpointRaw)
	if err != nil {
		h.Log().Error("error starting tx", "error", err)
		templates.Notice(
			templates.NewSiteNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("dashboard_invalid_slug"),
		).Render(ctx, w)
		return
	}

	for _, e := range config.Endpoints {
		used := strings.ReplaceAll(e, config.Endpoints[config.RootPath], "")
		used = strings.ReplaceAll(used, "/", "")

		if used == endpoint {
			h.Log().Debug("endpoint already in use by app", "slug", endpoint)
			templates.Notice(
				templates.NewSiteNoticeID,
				templates.NoticeWarn,
				tr("error"),
				tr("dashboard_slug_not_available"),
			).Render(ctx, w)
			return
		}
	}

	tx, err := h.DB().Begin()
	if err != nil {
		h.Log().Error("error starting tx", "error", err)
		templates.Notice(
			templates.NewSiteNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}
	defer tx.Rollback()

	queries := db.New(h.DB()).WithTx(tx)

	plan, err := queries.GetPlan(ctx, session.SessionUser)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		h.Log().Error("failed to query plan by user_id", "user", session.SessionUser)
		return
	}

	sites, err := queries.GetSitesWithMetricsByUserID(ctx, session.SessionUser)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		h.Log().Error("failed to query sites by user_id", "user", session.SessionUser)
		return
	}

	if len(sites)+1 > 1 {
		if plan.UserPlanActive != 1 || time.Now().Unix() > plan.UserPlanDueUnix {
			h.Log().Debug("tried to create a new site without required plan", "user", session.SessionUser)
			templates.Notice(
				templates.NewSiteNoticeID,
				templates.NoticeInfo,
				tr("info"),
				tr("dashboard_upgrade_to_create_more_sites"),
			).Render(ctx, w)
			return
		}

		if len(sites)+1 > 5 {
			h.Log().Debug("account reached max sites", "user", session.SessionUser)
			templates.Notice(
				templates.NewSiteNoticeID,
				templates.NoticeInfo,
				tr("info"),
				tr("dashboard_maximum_sites_reached"),
			).Render(ctx, w)
			return
		}
	}

	slugs, err := queries.GetSlugs(ctx)
	if err != nil {
		h.Log().Error("error querying all slugs", "error", err)
		templates.Notice(
			templates.NewSiteNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}

	if slices.Contains(slugs, endpoint) {
		h.Log().Debug("site slug/endpoint already exists", "slug", endpoint)
		templates.Notice(
			templates.NewSiteNoticeID,
			templates.NoticeWarn,
			tr("warn"),
			tr("dashboard_slug_not_available"),
		).Render(ctx, w)
		return
	}

	now := time.Now().Unix()

	siteID, err := queries.InsertSite(ctx, db.InsertSiteParams{
		SiteUser:          session.SessionUser,
		SiteSlug:          endpoint,
		SiteTitle:         name,
		SiteTagsJson:      "",
		SiteDescription:   "",
		SiteHtmlPublished: "",
		SiteCreatedUnix:   now,
		SiteModifiedUnix:  now,
		SitePublished:     0,
		SiteDeleted:       0,
		SiteHomePage:      0,
	})
	if err != nil {
		h.Log().Error("error inserting site", "site", endpoint, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := queries.InsertMetric(ctx, db.InsertMetricParams{
		MetricSite:        siteID,
		MetricVisitsTotal: 0,
	}); err != nil {
		h.Log().Error("error inserting metric", "site", endpoint, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		h.Log().Error("error tx commit", "site", endpoint, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := templates.Redirect(config.Endpoints[config.EditorPath]+endpoint).Render(ctx, w); err != nil {
		h.Log().Error("error rendering template", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func parseEndpoint(raw string) (string, error) {
	validEndpoint := regexp.MustCompile(`^[a-z0-9-]+$`)

	// Trim only leading/trailing whitespace
	endpoint := strings.TrimSpace(raw)

	// Must not be empty
	if endpoint == "" {
		return "", errors.New("endpoint cannot be empty")
	}

	// Must be all lowercase ASCII
	for _, r := range endpoint {
		if !unicode.IsLower(r) && !unicode.IsDigit(r) && r != '-' {
			return "", errors.New("endpoint must contain only lowercase letters, digits, or dashes")
		}
		if r > unicode.MaxASCII {
			return "", errors.New("endpoint must not contain accented or non-ASCII characters")
		}
	}

	// Validate full pattern
	if !validEndpoint.MatchString(endpoint) {
		return "", errors.New("endpoint must match ^[a-z0-9-]+$")
	}

	// Prevent accidental leading/trailing dash
	if strings.HasPrefix(endpoint, "-") || strings.HasSuffix(endpoint, "-") {
		return "", errors.New("endpoint cannot start or end with a dash")
	}

	// Check for consecutive dashes (optional but good hygiene)
	if strings.Contains(endpoint, "--") {
		return "", errors.New("endpoint cannot contain consecutive dashes")
	}

	// Reject if not valid UTF-8 (shouldn’t happen normally)
	if !utf8.ValidString(endpoint) {
		return "", errors.New("invalid UTF-8 in endpoint")
	}

	return endpoint, nil
}
