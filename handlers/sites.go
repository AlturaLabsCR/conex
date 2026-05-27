package handlers

import (
	"errors"
	"net/http"
	"strings"

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

	h.AddHandler(http.MethodGet, h.routePath("/sites/available/{path}"), http.HandlerFunc(h.SitePathAvailable))
	h.AddHandler(http.MethodPost, h.routePath("/sites"), authenticated(h.CreateSite))
	h.AddHandler(http.MethodPatch, h.routePath("/sites/{path}"), authenticated(h.SetSitePublic))
	h.AddHandler(http.MethodDelete, h.routePath("/sites/{path}"), authenticated(h.DeleteSite))
	h.AddHandler(http.MethodGet, h.routePath("/{path}"), http.HandlerFunc(h.GetSite))
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
			Subtitle:              site.Path,
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

func (h *Handler) CreateSite(w http.ResponseWriter, r *http.Request) {
	_, sub, status, ok := authenticatedAccountClaimsAndSubject(r)
	if !ok {
		h.writeStatus(w, r, status, "missing authenticated account subject")
		return
	}

	type createSiteRequest struct {
		Path string `json:"path"`
		HTML string `json:"html"`
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

	site, err := h.sites.Create(r.Context(), sub, req.Path, strings.NewReader(req.HTML))
	if err != nil {
		switch {
		case errors.Is(err, sites.ErrInvalidPath):
			h.writeError(w, r, http.StatusBadRequest, err, "invalid site path", "sub", sub, "site_path", req.Path)
			return
		case errors.Is(err, sites.ErrPathUnavailable):
			h.writeStatus(w, r, http.StatusConflict, "site path unavailable", "sub", sub, "site_path", req.Path)
			return
		default:
			h.writeError(w, r, http.StatusInternalServerError, err, "failed to create site", "sub", sub, "site_path", req.Path)
			return
		}
	}

	writeJSON(w, http.StatusCreated, site)
}

func (h *Handler) SetSitePublic(w http.ResponseWriter, r *http.Request) {
	_, sub, status, ok := authenticatedAccountClaimsAndSubject(r)
	if !ok {
		h.writeStatus(w, r, status, "missing authenticated account subject")
		return
	}

	type setSitePublicRequest struct {
		Public bool `json:"public"`
	}

	var req setSitePublicRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, err, "invalid site publish request body", "sub", sub, "site_path", r.PathValue("path"))
		return
	}

	site, err := h.sites.SetPublic(r.Context(), sub, r.PathValue("path"), req.Public)
	if err != nil {
		switch {
		case errors.Is(err, sites.ErrInvalidPath):
			h.writeError(w, r, http.StatusBadRequest, err, "invalid site path", "sub", sub, "site_path", r.PathValue("path"))
			return
		case errors.Is(err, sites.ErrSiteNotFound):
			h.writeError(w, r, http.StatusNotFound, err, "site not found", "sub", sub, "site_path", r.PathValue("path"))
			return
		default:
			h.writeError(w, r, http.StatusInternalServerError, err, "failed to update site public status", "sub", sub, "site_path", r.PathValue("path"))
			return
		}
	}

	writeJSON(w, http.StatusOK, site)
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
