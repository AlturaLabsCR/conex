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
	tr := h.Translator(r)

	// TODO:
	// - Check this site can be created given the user's permission

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
			h.Log().Error("endpoint already in use by app", "slug", endpoint)
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
		h.Log().Error("site slug/endpoint already exists", "slug", endpoint)
		templates.Notice(
			templates.NewSiteNoticeID,
			templates.NoticeWarn,
			tr("warn"),
			tr("dashboard_slug_not_available"),
		).Render(ctx, w)
		return
	}

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

	templates.Redirect(config.Endpoints[config.EditorPath]+endpoint).Render(ctx, w)
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
