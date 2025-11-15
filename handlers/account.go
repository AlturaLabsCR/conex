package handlers

import (
	"net/http"

	"app/internal/db"
	"app/templates"
)

func (h *Handler) Account(w http.ResponseWriter, r *http.Request) {
	h.Log().Debug("hit endpoint", "pattern", r.Pattern)

	ctx := r.Context()

	session, ok := ctx.Value(ctxSessionKey).(db.Session)
	if !ok {
		h.Log().Debug("error retrieving session from ctx")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	queries := db.New(h.DB())

	user, err := queries.GetUserByID(ctx, session.SessionUser)
	if err != nil {
		h.Log().Error("error retrieving user info")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sessions, err := queries.GetSessionsByUser(ctx, session.SessionUser)
	if err != nil {
		h.Log().Error("error retrieving user info")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	header := templates.AccountHeader(h.Translator(r))
	content := templates.Account(h.Translator(r), session, user, sessions)

	if err := templates.Base(h.Translator(r), header, content, true).Render(ctx, w); err != nil {
		h.Log().Error("error rendering template", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
