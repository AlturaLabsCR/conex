package handlers

import (
	"errors"
	"net/http"
	"strings"

	"app/database"
	"app/middleware"
	"app/sites"
	"app/templates/base"
	"app/templates/meta"
	sitetemplates "app/templates/sites"
)

func (h *Handler) registerSiteRoutes() {
	authenticated := func(fn http.HandlerFunc) http.Handler {
		return middleware.AuthenticateBearer(h.logger, h.authenticator, http.HandlerFunc(fn))
	}

	h.AddHandler(http.MethodGet, h.routePath("/{path}"), http.HandlerFunc(h.GetSite))
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
			h.writeError(w, r, http.StatusNotFound, err, "site not found", "site_path", r.PathValue("path"))
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
			Content: sitetemplates.SiteMain(html),
			Active:  h.routePath("/" + site.Path),
		},
	})

	if err := page.Render(r.Context(), w); err != nil {
		h.writeError(w, r, http.StatusInternalServerError, err, "failed to render site", "site_path", site.Path)
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
		Sub    int64    `json:"sub"`
		Path   string   `json:"path"`
		Public bool     `json:"public"`
		Name   string   `json:"name"`
		Tags   []string `json:"tags"`
		URL    string   `json:"url"`
		HTML   string   `json:"html"`
	}

	base := h.siteResponse(*site)
	writeJSON(w, http.StatusOK, ownedSiteResponse{
		Sub:    base.Sub,
		Path:   base.Path,
		Public: base.Public,
		Name:   base.Name,
		Tags:   base.Tags,
		URL:    base.URL,
		HTML:   html,
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
		default:
			h.writeError(w, r, http.StatusInternalServerError, err, "failed to update site", "sub", sub, "site_path", r.PathValue("path"))
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

type siteResponse struct {
	Sub    int64    `json:"sub"`
	Path   string   `json:"path"`
	Public bool     `json:"public"`
	Name   string   `json:"name"`
	Tags   []string `json:"tags"`
	URL    string   `json:"url"`
}

func (h *Handler) siteResponse(site database.Site) siteResponse {
	return siteResponse{
		Sub:    site.Sub,
		Path:   site.Path,
		Public: site.Public,
		Name:   site.Name,
		Tags:   site.Tags,
		URL:    h.siteURL(site.Path),
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
