package handlers

import (
	"net/http"

	"app/config"
	"app/internal/db"
	"app/templates"
)

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// TODO: if not logged in, redirect to login page
	if false {
		http.Redirect(w, r, config.LoginPath, http.StatusSeeOther)
	}

	header := templates.DashboardHeader(h.Translator(r))
	content := templates.CardsGrid(h.Translator(r), []db.ValidSitesWithMetric{})

	templates.Base(h.Translator(r), header, content).Render(ctx, w)
}
