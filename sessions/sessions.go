// Package sessions implements typed JWT cookie sessions.
package sessions

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims holds typed session data plus standard JWT claims.
type Claims[T any] struct {
	Data T
	jwt.RegisteredClaims
}

// Store handles creating and validating JWT-backed cookies.
type Store[T any] struct {
	params StoreParams
}

type StoreParams struct {
	CookieName     string
	CookiePath     string
	CookieSameSite http.SameSite
	CookieTTL      time.Duration

	JWTSecret string
}

// New creates a new typed session store.
func New[T any](params StoreParams) *Store[T] {
	return &Store[T]{params: params}
}

// JWTSet creates and signs a JWT, then stores it in a secure HttpOnly cookie.
func (s *Store[T]) JWTSet(w http.ResponseWriter, r *http.Request, data T) (string, error) {
	expires := time.Now().Add(s.params.CookieTTL)

	claims := &Claims[T]{
		Data: data,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expires),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	tokenStr, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(s.params.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     s.params.CookieName,
		Path:     s.params.CookiePath,
		SameSite: s.params.CookieSameSite,
		Expires:  expires,
		Value:    tokenStr,
		HttpOnly: true,
		Secure:   r.TLS != nil,
	})

	return tokenStr, nil
}

// JWTValidate reads and verifies the cookie, returning the typed session data.
func (s *Store[T]) JWTValidate(r *http.Request) (T, error) {
	var zero T // zero value of T if validation fails

	cookie, err := r.Cookie(s.params.CookieName)
	if err != nil {
		return zero, fmt.Errorf("get cookie: %w", err)
	}

	if !cookie.Expires.IsZero() && cookie.Expires.Before(time.Now()) {
		return zero, fmt.Errorf("cookie expired at %v", cookie.Expires)
	}

	token, err := jwt.ParseWithClaims(cookie.Value, &Claims[T]{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.params.JWTSecret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return zero, fmt.Errorf("token expired: %w", err)
		}
		return zero, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims[T])
	if !ok || !token.Valid {
		return zero, fmt.Errorf("invalid token claims")
	}

	return claims.Data, nil
}
