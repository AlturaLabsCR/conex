package handlers

import (
	"net/http"
	"time"

	"app/database"
	"app/middleware"
)

func (h *Handler) registerAccountRoutes() {
	authenticated := func(fn http.HandlerFunc) http.Handler {
		return middleware.AuthenticateBearer(h.logger, h.authenticator, h.localizeError, http.HandlerFunc(fn))
	}

	h.AddHandler(http.MethodGet, h.routePath("/api/account"), authenticated(h.GetAccount))
	h.AddHandler(http.MethodDelete, h.routePath("/api/account"), authenticated(h.DeleteAccount))
	h.AddHandler(http.MethodPatch, h.routePath("/api/account/email/change"), authenticated(h.RequestEmailChange))
	h.AddHandler(http.MethodPatch, h.routePath("/api/account/email/change/confirm"), authenticated(h.ConfirmEmailChange))
}

func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	_, sub, status, ok := authenticatedAccountClaimsAndSubject(r)
	if !ok {
		h.writeStatus(w, r, status, "missing authenticated account subject")
		return
	}

	account, err := h.db.Querier().SelectAccountBySub(r.Context(), sub)
	if err != nil {
		if h.db.IsErrNotFound(err) {
			h.writeError(w, r, http.StatusNotFound, err, "account not found", "sub", sub)
			return
		}

		h.writeError(w, r, http.StatusInternalServerError, err, "failed to select account", "sub", sub)
		return
	}

	subscription, err := h.db.Querier().SelectAccountSubscriptionBySub(r.Context(), sub)
	if err != nil {
		if h.db.IsErrNotFound(err) {
			h.writeError(w, r, http.StatusNotFound, err, "account subscription not found", "sub", sub)
			return
		}

		h.writeError(w, r, http.StatusInternalServerError, err, "failed to select account subscription", "sub", sub)
		return
	}

	L := h.localizer.LocalizerFunc(h.localizer.PickLanguageFromRequest(r))

	type subscriptionResponse struct {
		Status  string       `json:"status"`
		DueDate string       `json:"due_date"`
		Plan    planResponse `json:"plan"`
	}

	type accountResponse struct {
		Sub          int64                `json:"sub"`
		Email        string               `json:"email"`
		CreatedAt    int64                `json:"created_at"`
		Subscription subscriptionResponse `json:"subscription"`
	}

	writeJSON(w, http.StatusOK, accountResponse{
		Sub:       account.Sub,
		Email:     account.Email,
		CreatedAt: account.CreatedAt,
		Subscription: subscriptionResponse{
			Status:  subscription.Status,
			DueDate: subscription.DueDate,
			Plan: h.planResponse(L, database.Plan{
				ID:              subscription.PlanID,
				NameKey:         subscription.PlanNameKey,
				PriceAmount:     subscription.PriceAmount,
				PriceCurrency:   subscription.PriceCurrency,
				BillingUnit:     subscription.BillingUnit,
				BillingCount:    subscription.BillingCount,
				SupportsRenewal: subscription.SupportsRenewal,
				Policy:          subscription.Policy,
			}),
		},
	})
}

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	identity, sub, status, ok := authenticatedAccountClaimsAndSubject(r)
	if !ok {
		h.writeStatus(w, r, status, "missing authenticated account subject")
		return
	}

	if err := h.authenticator.RevokeAll(r.Context(), identity); err != nil {
		h.writeError(w, r, http.StatusInternalServerError, err, "failed to revoke account sessions", "sub", sub)
		return
	}

	if err := h.sites.DeleteAll(r.Context(), sub); err != nil {
		h.writeError(w, r, http.StatusInternalServerError, err, "failed to delete account sites", "sub", sub)
		return
	}

	if err := h.db.WithTx(r.Context(), func(q database.Querier) error {
		if err := q.DeleteAccountEmailChangeRequest(r.Context(), sub); err != nil {
			return err
		}

		return q.DeleteAccount(r.Context(), sub)
	}); err != nil {
		h.writeError(w, r, http.StatusInternalServerError, err, "failed to delete account", "sub", sub)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RequestEmailChange(w http.ResponseWriter, r *http.Request) {
	_, sub, status, ok := authenticatedAccountClaimsAndSubject(r)
	if !ok {
		h.writeStatus(w, r, status, "missing authenticated account subject")
		return
	}

	L := h.localizer.LocalizerFunc(h.localizer.PickLanguageFromRequest(r))

	type changeEmailRequest struct {
		NewEmail string `json:"new_email"`
	}

	var req changeEmailRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, err, "invalid account email change request body")
		return
	}

	newEmail, err := validEmail(req.NewEmail)
	if err != nil {
		h.writeStatus(w, r, http.StatusBadRequest, "invalid account email change target", "email", req.NewEmail, "sub", sub)
		return
	}

	account, err := h.db.Querier().SelectAccountBySub(r.Context(), sub)
	if err != nil {
		if h.db.IsErrNotFound(err) {
			h.writeError(w, r, http.StatusNotFound, err, "account not found", "sub", sub)
			return
		}

		h.writeError(w, r, http.StatusInternalServerError, err, "failed to select account", "sub", sub)
		return
	}

	if account.Email == newEmail {
		h.writeStatus(w, r, http.StatusBadRequest, "account email change target matches current email", "sub", sub, "email", newEmail)
		return
	}

	otp, err := randomOTP()
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, err, "failed to generate account email change otp", "sub", sub)
		return
	}

	expiresAt := time.Now().UTC().Add(10 * time.Minute).Unix()
	if err := h.db.Querier().UpsertAccountEmailChangeRequest(r.Context(), sub, newEmail, otp, expiresAt); err != nil {
		h.writeError(w, r, http.StatusInternalServerError, err, "failed to upsert account email change request", "sub", sub, "email", newEmail)
		return
	}

	if h.dev {
		h.logger.Debug("otp sent", "new_email", newEmail, "otp", otp)
	} else {
		if err := h.mailer.SendOTP(r.Context(), L, newEmail, otp, expiresAt); err != nil {
			h.writeError(w, r, http.StatusInternalServerError, err, "failed to send account email change otp", "sub", sub, "email", newEmail, "expires_at", expiresAt)
			return
		}
		h.logger.Debug("otp sent", "email", newEmail)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ConfirmEmailChange(w http.ResponseWriter, r *http.Request) {
	_, sub, status, ok := authenticatedAccountClaimsAndSubject(r)
	if !ok {
		h.writeStatus(w, r, status, "missing authenticated account subject")
		return
	}

	type changeEmailConfirmRequest struct {
		OTP int64 `json:"otp"`
	}

	var req changeEmailConfirmRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, err, "invalid account email confirm request body")
		return
	}

	saved, err := h.db.Querier().SelectAccountEmailChangeRequestBySub(r.Context(), sub)
	if err != nil {
		if h.db.IsErrNotFound(err) {
			h.writeError(w, r, http.StatusNotFound, err, "missing account email change request", "sub", sub)
			return
		}

		h.writeError(w, r, http.StatusInternalServerError, err, "failed to select account email change request", "sub", sub)
		return
	}

	if saved.Otp != req.OTP || saved.ExpiresAt < time.Now().UTC().Unix() {
		if saved.ExpiresAt < time.Now().UTC().Unix() {
			_ = h.db.Querier().DeleteAccountEmailChangeRequest(r.Context(), sub)
		}

		h.writeStatus(w, r, http.StatusBadRequest, "invalid account email change verification code", "sub", sub)
		return
	}

	if err := h.db.WithTx(r.Context(), func(q database.Querier) error {
		if err := q.DeleteAccountEmailChangeRequest(r.Context(), sub); err != nil {
			return err
		}

		return q.UpdateAccountEmail(r.Context(), sub, saved.Email)
	}); err != nil {
		h.writeError(w, r, http.StatusInternalServerError, err, "failed to confirm account email change", "sub", sub, "email", saved.Email)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
