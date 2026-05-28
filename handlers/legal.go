package handlers

import (
	"net/http"

	"app/templates/base"
	"app/templates/legal"
	"app/templates/meta"
)

func (h *Handler) registerLegalRoutes() {
	h.Add(http.MethodGet, h.routePath("/tos"), h.TermsOfService)
	h.Add(http.MethodGet, h.routePath("/privacy"), h.PrivacyPolicy)
}

func (h *Handler) TermsOfService(w http.ResponseWriter, r *http.Request) {
	L := h.localizer.LocalizerFunc(h.localizer.PickLanguageFromRequest(r))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page := base.Page(L, base.PageParams{
		Head: base.HeadParams{
			Title:       meta.AppTitle,
			Subtitle:    "Terms of Service",
			Description: "Terms of Service for Conex.",
			RobotsIndex: true,
		},
		Body: base.BodyParams{
			Content: legal.TermsOfService(),
			Active:  h.routePath("/tos"),
		},
	})

	if err := page.Render(r.Context(), w); err != nil {
		h.writeError(w, r, http.StatusInternalServerError, err, "failed to render terms of service")
	}
}

func (h *Handler) PrivacyPolicy(w http.ResponseWriter, r *http.Request) {
	L := h.localizer.LocalizerFunc(h.localizer.PickLanguageFromRequest(r))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page := base.Page(L, base.PageParams{
		Head: base.HeadParams{
			Title:       meta.AppTitle,
			Subtitle:    "Privacy Policy",
			Description: "Privacy Policy for Conex.",
			RobotsIndex: true,
		},
		Body: base.BodyParams{
			Content: legal.PrivacyPolicy(),
			Active:  h.routePath("/privacy"),
		},
	})

	if err := page.Render(r.Context(), w); err != nil {
		h.writeError(w, r, http.StatusInternalServerError, err, "failed to render privacy policy")
	}
}
