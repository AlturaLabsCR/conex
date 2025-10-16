package middleware

import (
	"net/http"
)

func DisableCacheInDevMode(production bool, next http.Handler) http.Handler {
	if !production {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
