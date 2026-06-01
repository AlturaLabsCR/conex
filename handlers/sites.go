package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"app/database"
	"app/middleware"
	"app/sites"
	"app/templates/base"
	"app/templates/meta"
	sitetemplates "app/templates/sites"
)

const maxSiteListPage = 107374183

func (h *Handler) registerSiteRoutes() {
	authenticated := func(fn http.HandlerFunc) http.Handler {
		return middleware.AuthenticateBearer(h.logger, h.authenticator, h.localizeError, http.HandlerFunc(fn))
	}

	h.AddHandler(http.MethodGet, h.routePath("/{path}"), http.HandlerFunc(h.GetSite))
	h.AddHandler(http.MethodGet, h.routePath("/api/public/sites/top"), http.HandlerFunc(h.ListTopSites))
	h.AddHandler(http.MethodGet, h.routePath("/api/public/sites/top/{page}"), http.HandlerFunc(h.ListTopSites))
	h.AddHandler(http.MethodGet, h.routePath("/api/public/sites/latest"), http.HandlerFunc(h.ListLatestSites))
	h.AddHandler(http.MethodGet, h.routePath("/api/public/sites/latest/{page}"), http.HandlerFunc(h.ListLatestSites))
	h.AddHandler(http.MethodGet, h.routePath("/api/public/sites/search/{query}"), http.HandlerFunc(h.SearchSites))
	h.AddHandler(http.MethodGet, h.routePath("/api/public/sites/search/{query}/{page}"), http.HandlerFunc(h.SearchSites))
	h.AddHandler(http.MethodHead, h.routePath("/api/sites/{path}"), http.HandlerFunc(h.SitePathAvailable))
	h.AddHandler(http.MethodPost, h.routePath("/api/sites"), authenticated(h.CreateSite))
	h.AddHandler(http.MethodGet, h.routePath("/api/sites"), authenticated(h.ListSites))
	h.AddHandler(http.MethodGet, h.routePath("/api/sites/{path}"), authenticated(h.GetOwnedSite))
	h.AddHandler(http.MethodPatch, h.routePath("/api/sites/{path}"), authenticated(h.UpdateSite))
	h.AddHandler(http.MethodDelete, h.routePath("/api/sites/{path}"), authenticated(h.DeleteSite))
}

func (h *Handler) SitePathAvailable(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path")
	if h.sitePathConflicts(path) {
		h.writeStatus(w, r, http.StatusConflict, "site path conflicts with registered route", "site_path", path)
		return
	}

	available, err := h.sites.IsPathAvailable(r.Context(), path)
	if err != nil {
		if errors.Is(err, sites.ErrInvalidPath) {
			h.writeError(w, r, http.StatusBadRequest, err, "invalid site path", "site_path", path)
			return
		}

		h.writeError(w, r, http.StatusInternalServerError, err, "failed to check site path availability", "site_path", path)
		return
	}
	if !available {
		h.writeStatus(w, r, http.StatusConflict, "site path unavailable", "site_path", path)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetSite(w http.ResponseWriter, r *http.Request) {
	site, html, err := h.sites.Get(r.Context(), r.PathValue("path"))
	if err != nil {
		switch {
		case errors.Is(err, sites.ErrInvalidPath):
			h.writeError(w, r, http.StatusBadRequest, err, "invalid site path", "site_path", r.PathValue("path"))
			return
		case errors.Is(err, sites.ErrSiteNotFound):
			h.renderSiteNotFound(w, r, err)
			return
		default:
			h.writeError(w, r, http.StatusInternalServerError, err, "failed to get site", "site_path", r.PathValue("path"))
			return
		}
	}

	L := h.localizer.LocalizerFunc(h.localizer.PickLanguageFromRequest(r))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page := base.Page(L, base.PageParams{
		Head: base.HeadParams{
			Title:                 meta.AppTitle,
			Subtitle:              site.Name,
			RobotsIndex:           true,
			RobotsGoogleTranslate: true,
		},
		Body: base.BodyParams{
			Content:       sitetemplates.SiteMain(html),
			Active:        h.routePath("/" + site.Path),
			HeaderTitle:   site.Name,
			PoweredFooter: true,
		},
	})

	if err := page.Render(r.Context(), w); err != nil {
		h.writeError(w, r, http.StatusInternalServerError, err, "failed to render site", "site_path", site.Path)
	}
}

func (h *Handler) renderSiteNotFound(w http.ResponseWriter, r *http.Request, err error) {
	path := r.PathValue("path")
	h.logger.Debug("site not found", "status", http.StatusNotFound, "method", r.Method, "path", r.URL.Path, "error", err, "site_path", path)

	L := h.localizer.LocalizerFunc(h.localizer.PickLanguageFromRequest(r))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)

	page := base.Page(L, base.PageParams{
		Head: base.HeadParams{
			Title:                 meta.AppTitle,
			Subtitle:              L("site not found"),
			RobotsGoogleTranslate: true,
		},
		Body: base.BodyParams{
			Content: sitetemplates.SiteNotFound(L),
			Active:  h.routePath("/" + path),
		},
	})

	if renderErr := page.Render(r.Context(), w); renderErr != nil {
		h.logger.Error("failed to render site not found page", "status", http.StatusInternalServerError, "method", r.Method, "path", r.URL.Path, "error", renderErr, "site_path", path)
	}
}

func (h *Handler) ListSites(w http.ResponseWriter, r *http.Request) {
	_, sub, status, ok := authenticatedAccountClaimsAndSubject(r)
	if !ok {
		h.writeStatus(w, r, status, "missing authenticated account subject")
		return
	}

	ownedSites, err := h.sites.List(r.Context(), sub)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, err, "failed to list sites", "sub", sub)
		return
	}

	writeJSON(w, http.StatusOK, h.siteResponses(ownedSites))
}

func (h *Handler) ListTopSites(w http.ResponseWriter, r *http.Request) {
	page, ok := h.siteListPage(w, r)
	if !ok {
		return
	}

	publicSites, err := h.sites.ListTop(r.Context(), page)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, err, "failed to list top sites", "page", page)
		return
	}

	writeJSON(w, http.StatusOK, h.siteResponses(publicSites))
}

func (h *Handler) ListLatestSites(w http.ResponseWriter, r *http.Request) {
	page, ok := h.siteListPage(w, r)
	if !ok {
		return
	}

	publicSites, err := h.sites.ListLatest(r.Context(), page)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, err, "failed to list latest sites", "page", page)
		return
	}

	writeJSON(w, http.StatusOK, h.siteResponses(publicSites))
}

func (h *Handler) SearchSites(w http.ResponseWriter, r *http.Request) {
	page, ok := h.siteListPage(w, r)
	if !ok {
		return
	}

	query := strings.TrimSpace(r.PathValue("query"))
	publicSites, err := h.sites.Search(r.Context(), query, page)
	if err != nil {
		switch {
		case errors.Is(err, sites.ErrInvalidName):
			h.writeError(w, r, http.StatusBadRequest, err, "invalid site search query", "query", query, "page", page)
			return
		default:
			h.writeError(w, r, http.StatusInternalServerError, err, "failed to search sites", "query", query, "page", page)
			return
		}
	}

	writeJSON(w, http.StatusOK, h.siteResponses(publicSites))
}

func (h *Handler) siteListPage(w http.ResponseWriter, r *http.Request) (int64, bool) {
	raw := strings.TrimSpace(r.PathValue("page"))
	if raw == "" {
		return 1, true
	}

	page, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || page < 1 || page > maxSiteListPage {
		h.writeStatus(w, r, http.StatusBadRequest, "invalid site list page", "page", raw)
		return 0, false
	}

	return page, true
}

func (h *Handler) GetOwnedSite(w http.ResponseWriter, r *http.Request) {
	_, sub, status, ok := authenticatedAccountClaimsAndSubject(r)
	if !ok {
		h.writeStatus(w, r, status, "missing authenticated account subject")
		return
	}

	site, html, err := h.sites.GetOwned(r.Context(), sub, r.PathValue("path"))
	if err != nil {
		switch {
		case errors.Is(err, sites.ErrInvalidPath):
			h.writeError(w, r, http.StatusBadRequest, err, "invalid site path", "sub", sub, "site_path", r.PathValue("path"))
			return
		case errors.Is(err, sites.ErrSiteNotFound):
			h.writeError(w, r, http.StatusNotFound, err, "site not found", "sub", sub, "site_path", r.PathValue("path"))
			return
		default:
			h.writeError(w, r, http.StatusInternalServerError, err, "failed to get owned site", "sub", sub, "site_path", r.PathValue("path"))
			return
		}
	}

	type ownedSiteResponse struct {
		Sub          int64    `json:"sub"`
		Path         string   `json:"path"`
		Public       bool     `json:"public"`
		Name         string   `json:"name"`
		Tags         []string `json:"tags"`
		URL          string   `json:"url"`
		CreatedAt    int64    `json:"created_at"`
		LastModified int64    `json:"last_modified"`
		Clicks       int64    `json:"clicks"`
		HTML         string   `json:"html"`
	}

	base := h.siteResponse(*site)
	writeJSON(w, http.StatusOK, ownedSiteResponse{
		Sub:          base.Sub,
		Path:         base.Path,
		Public:       base.Public,
		Name:         base.Name,
		Tags:         base.Tags,
		URL:          base.URL,
		CreatedAt:    base.CreatedAt,
		LastModified: base.LastModified,
		Clicks:       base.Clicks,
		HTML:         html,
	})
}

func (h *Handler) CreateSite(w http.ResponseWriter, r *http.Request) {
	_, sub, status, ok := authenticatedAccountClaimsAndSubject(r)
	if !ok {
		h.writeStatus(w, r, status, "missing authenticated account subject")
		return
	}

	type createSiteRequest struct {
		Path string   `json:"path"`
		Name string   `json:"name"`
		Tags []string `json:"tags"`
		HTML string   `json:"html"`
	}

	var req createSiteRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, err, "invalid site create request body", "sub", sub)
		return
	}

	if h.sitePathConflicts(req.Path) {
		h.writeStatus(w, r, http.StatusConflict, "site path conflicts with registered route", "sub", sub, "site_path", req.Path)
		return
	}

	site, err := h.sites.Create(r.Context(), sub, req.Path, req.Name, req.Tags, strings.NewReader(req.HTML))
	if err != nil {
		switch {
		case errors.Is(err, sites.ErrInvalidPath):
			h.writeError(w, r, http.StatusBadRequest, err, "invalid site path", "sub", sub, "site_path", req.Path)
			return
		case errors.Is(err, sites.ErrInvalidName):
			h.writeError(w, r, http.StatusBadRequest, err, "invalid site name", "sub", sub, "site_path", req.Path)
			return
		case errors.Is(err, sites.ErrInvalidTags):
			h.writeError(w, r, http.StatusBadRequest, err, "invalid site tags", "sub", sub, "site_path", req.Path)
			return
		case errors.Is(err, sites.ErrPathUnavailable):
			h.writeStatus(w, r, http.StatusConflict, "site path unavailable", "sub", sub, "site_path", req.Path)
			return
		case errors.Is(err, sites.ErrSiteSizeLimit):
			h.writeError(w, r, http.StatusRequestEntityTooLarge, err, "site size limit exceeded", "sub", sub, "site_path", req.Path)
			return
		case errors.Is(err, sites.ErrSiteCountLimit):
			h.writeError(w, r, http.StatusForbidden, err, "site count limit exceeded", "sub", sub, "site_path", req.Path)
			return
		case errors.Is(err, sites.ErrSubpathLimit):
			h.writeError(w, r, http.StatusForbidden, err, "site subpath limit exceeded", "sub", sub, "site_path", req.Path)
			return
		default:
			h.writeError(w, r, http.StatusInternalServerError, err, "failed to create site", "sub", sub, "site_path", req.Path)
			return
		}
	}

	writeJSON(w, http.StatusCreated, h.siteResponse(*site))
}

func (h *Handler) UpdateSite(w http.ResponseWriter, r *http.Request) {
	_, sub, status, ok := authenticatedAccountClaimsAndSubject(r)
	if !ok {
		h.writeStatus(w, r, status, "missing authenticated account subject")
		return
	}

	type updateSiteRequest struct {
		Public *bool     `json:"public"`
		Name   *string   `json:"name"`
		Tags   *[]string `json:"tags"`
		HTML   *string   `json:"html"`
	}

	var req updateSiteRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, err, "invalid site update request body", "sub", sub, "site_path", r.PathValue("path"))
		return
	}

	_, err := h.sites.Update(r.Context(), sub, r.PathValue("path"), sites.SiteUpdate{
		Public: req.Public,
		Name:   req.Name,
		Tags:   req.Tags,
		HTML:   req.HTML,
	})
	if err != nil {
		switch {
		case errors.Is(err, sites.ErrInvalidPath):
			h.writeError(w, r, http.StatusBadRequest, err, "invalid site path", "sub", sub, "site_path", r.PathValue("path"))
			return
		case errors.Is(err, sites.ErrInvalidName):
			h.writeError(w, r, http.StatusBadRequest, err, "invalid site name", "sub", sub, "site_path", r.PathValue("path"))
			return
		case errors.Is(err, sites.ErrInvalidTags):
			h.writeError(w, r, http.StatusBadRequest, err, "invalid site tags", "sub", sub, "site_path", r.PathValue("path"))
			return
		case errors.Is(err, sites.ErrSiteNotFound):
			h.writeError(w, r, http.StatusNotFound, err, "site not found", "sub", sub, "site_path", r.PathValue("path"))
			return
		case errors.Is(err, sites.ErrSiteSizeLimit):
			h.writeError(w, r, http.StatusRequestEntityTooLarge, err, "site size limit exceeded", "sub", sub, "site_path", r.PathValue("path"))
			return
		default:
			h.writeError(w, r, http.StatusInternalServerError, err, "failed to update site", "sub", sub, "site_path", r.PathValue("path"))
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

type siteResponse struct {
	Sub          int64    `json:"sub"`
	Path         string   `json:"path"`
	Public       bool     `json:"public"`
	Name         string   `json:"name"`
	Tags         []string `json:"tags"`
	URL          string   `json:"url"`
	CreatedAt    int64    `json:"created_at"`
	LastModified int64    `json:"last_modified"`
	Clicks       int64    `json:"clicks"`
}

func (h *Handler) siteResponse(site database.Site) siteResponse {
	return siteResponse{
		Sub:          site.Sub,
		Path:         site.Path,
		Public:       site.Public,
		Name:         site.Name,
		Tags:         site.Tags,
		URL:          h.siteURL(site.Path),
		CreatedAt:    site.CreatedAt,
		LastModified: site.LastModified,
		Clicks:       site.Clicks,
	}
}

func (h *Handler) siteResponses(sites []database.Site) []siteResponse {
	out := make([]siteResponse, 0, len(sites))
	for _, site := range sites {
		out = append(out, h.siteResponse(site))
	}

	return out
}

func (h *Handler) DeleteSite(w http.ResponseWriter, r *http.Request) {
	_, sub, status, ok := authenticatedAccountClaimsAndSubject(r)
	if !ok {
		h.writeStatus(w, r, status, "missing authenticated account subject")
		return
	}

	if err := h.sites.Delete(r.Context(), sub, r.PathValue("path")); err != nil {
		switch {
		case errors.Is(err, sites.ErrInvalidPath):
			h.writeError(w, r, http.StatusBadRequest, err, "invalid site path", "sub", sub, "site_path", r.PathValue("path"))
			return
		case errors.Is(err, sites.ErrSiteNotFound):
			h.writeError(w, r, http.StatusNotFound, err, "site not found", "sub", sub, "site_path", r.PathValue("path"))
			return
		default:
			h.writeError(w, r, http.StatusInternalServerError, err, "failed to delete site", "sub", sub, "site_path", r.PathValue("path"))
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) sitePathConflicts(path string) bool {
	path = strings.TrimSpace(strings.Trim(path, "/"))
	if path == "" {
		return true
	}

	candidate := h.routePath("/" + path)
	for _, registeredPath := range h.paths {
		if registeredPath == candidate {
			return true
		}

		if strings.HasPrefix(registeredPath, candidate+"/") || strings.HasPrefix(candidate, registeredPath+"/") {
			return true
		}
	}

	return false
}
