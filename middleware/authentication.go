// Package middleware provides HTTP middleware helpers.
package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	appauth "app/auth"
	goauth "github.com/tavocg/go-auth"
)

type authenticatedClaimsContextKey struct{}

var AuthenticatedClaimsContextKey = authenticatedClaimsContextKey{}

const StatusReauthenticationRequired = http.StatusUnauthorized

type ErrorLocalizer func(*http.Request, string) string

type errorResponse struct {
	Error errorResponseBody `json:"error"`
}

type errorResponseBody struct {
	Msg string `json:"msg"`
}

func AuthenticateBearer(logger Logger, authenticator goauth.Authenticator[*appauth.Claims], localizeError ErrorLocalizer, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Fields(strings.TrimSpace(r.Header.Get("Authorization")))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			logger.Debug("missing bearer token", "status", StatusReauthenticationRequired, "method", r.Method, "path", r.URL.Path)
			writeReauthenticationRequired(w, r, "missing bearer token", localizeError)
			return
		}

		identity, err := authenticator.Verify(r.Context(), parts[1])
		if err != nil {
			if errors.Is(err, goauth.ErrInvalidToken) || errors.Is(err, goauth.ErrExpiredToken) {
				logger.Debug("failed to verify bearer token", "status", StatusReauthenticationRequired, "method", r.Method, "path", r.URL.Path, "error", err)
				writeReauthenticationRequired(w, r, "failed to verify bearer token", localizeError)
				return
			}

			logger.Error("failed to verify bearer token", "status", http.StatusInternalServerError, "method", r.Method, "path", r.URL.Path, "error", err)
			writeError(w, r, http.StatusInternalServerError, "failed to verify bearer token", localizeError)
			return
		}

		ctx := context.WithValue(r.Context(), AuthenticatedClaimsContextKey, identity)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeReauthenticationRequired(w http.ResponseWriter, r *http.Request, msg string, localize ErrorLocalizer) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	writeError(w, r, StatusReauthenticationRequired, msg, localize)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, msg string, localize ErrorLocalizer) {
	if localize != nil {
		msg = localize(r, msg)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{
		Error: errorResponseBody{
			Msg: msg,
		},
	})
}

func AuthenticatedClaims(ctx context.Context) (*appauth.Claims, bool) {
	identity, ok := ctx.Value(AuthenticatedClaimsContextKey).(*appauth.Claims)
	return identity, ok
}
