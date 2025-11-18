package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"app/config"
	"app/internal/db"
	"app/templates"
	"app/utils"
)

func (h *Handler) Pricing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	session, ok := ctx.Value(ctxSessionKey).(db.Session)
	if !ok {
		h.Log().Error("error retrieving session from ctx")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	now := time.Now().Unix()

	var plan db.UserPlan
	var err error = nil

	plan, err = h.Queries().GetPlan(ctx, session.SessionUser)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			h.Log().Error("error retrieving plan", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	tr := h.Translator(r)

	subscribed := plan.UserPlanActive == 1 && now < plan.UserPlanDueUnix

	header := templates.PricingHeader(tr)
	content := templates.Pricing(tr, config.PayPalClientID, subscribed, utils.UnixToYMD(plan.UserPlanDueUnix))

	if err := templates.Base(tr, header, content, nil, true).Render(ctx, w); err != nil {
		h.Log().Error("error rendering template", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
