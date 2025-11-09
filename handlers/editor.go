package handlers

import (
	"net/http"

	"app/internal/db"
	"app/templates"
)

func (h *Handler) Editor(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	s := r.PathValue("site")

	queries := db.New(h.DB())

	site, err := queries.GetSiteWithMetrics(ctx, s)
	if err != nil {
		h.Log().Error("error loading sites", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	header := templates.EditorHeader(h.Translator((r)), site)
	content := templates.Editor(site)

	templates.Base(h.Translator(r), header, content).Render(ctx, w)
}
