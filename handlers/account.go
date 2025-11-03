package handlers

import (
	"net/http"

	"app/internal/db"
	"app/templates"
)

func (h *Handler) Account(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	session, ok := ctx.Value(ctxSessionKey).(db.Session)
	if !ok {
		h.Log().Error("error retrieving session from ctx")
		return
	}

	queries := db.New(h.DB())

	user, err := queries.GetUserByID(ctx, session.SessionUser)
	if err != nil {
		h.Log().Error("error retrieving user info")
		return
	}

	sessions, err := queries.GetSessionsByUser(ctx, session.SessionUser)
	if err != nil {
		h.Log().Error("error retrieving user info")
		return
	}

	header := templates.AccountHeader(h.Translator(r))
	content := templates.Account(h.Translator(r), session, user, sessions)

	templates.Base(h.Translator(r), header, content).Render(ctx, w)
}
