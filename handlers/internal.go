package handlers

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"

	appauth "app/auth"
	"app/middleware"

	"github.com/tavocg/go-email"
)

func decodeJSON(body io.Reader, dst any) error {
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("request body must contain a single JSON value")
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) writeStatus(w http.ResponseWriter, r *http.Request, status int, msg string, args ...any) {
	logArgs := []any{
		"status", status,
		"method", r.Method,
		"path", r.URL.Path,
	}
	logArgs = append(logArgs, args...)

	if status >= http.StatusInternalServerError {
		h.logger.Error(msg, logArgs...)
	} else if status >= http.StatusBadRequest {
		h.logger.Debug(msg, logArgs...)
	}

	w.WriteHeader(status)
}

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, status int, err error, msg string, args ...any) {
	logArgs := []any{
		"status", status,
		"method", r.Method,
		"path", r.URL.Path,
	}
	if err != nil {
		logArgs = append(logArgs, "error", err)
	}
	logArgs = append(logArgs, args...)

	if status >= http.StatusInternalServerError {
		h.logger.Error(msg, logArgs...)
	} else if status >= http.StatusBadRequest {
		h.logger.Debug(msg, logArgs...)
	}

	w.WriteHeader(status)
}

func authenticatedAccountClaimsAndSubject(r *http.Request) (*appauth.Claims, int64, int, bool) {
	identity, ok := middleware.AuthenticatedClaims(r.Context())
	if !ok {
		return nil, 0, http.StatusInternalServerError, false
	}

	if identity == nil || identity.Sub == "" {
		return nil, 0, http.StatusUnauthorized, false
	}

	sub, err := strconv.ParseInt(identity.Sub, 10, 64)
	if err != nil || sub <= 0 {
		return nil, 0, http.StatusUnauthorized, false
	}

	return identity, sub, http.StatusOK, true
}

func randomOTP() (int64, error) {
	max := big.NewInt(1_000_000)
	value, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}

	return value.Int64(), nil
}

func validEmail(address string) (string, error) {
	valid, err := email.StrictParser(address)
	if err != nil || valid.IsBlacklisted() {
		return "", fmt.Errorf("invalid email address")
	}
	valid.Normalize()
	return valid.Address(), nil
}
