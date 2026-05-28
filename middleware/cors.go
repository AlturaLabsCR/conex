// Package middleware provides HTTP middleware helpers.
package middleware

import (
	"net/http"
	"net/url"
	"slices"
	"strings"
)

type CORSOptions struct {
	AllowedOrigins []string
}

func CORS(opts CORSOptions, next http.Handler) http.Handler {
	allowedOrigins := cleanOrigins(opts.AllowedOrigins)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin != "" && originAllowed(origin, allowedOrigins) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Accept-Language, Authorization, Content-Type")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func cleanOrigins(origins []string) []string {
	cleaned := []string{}
	for _, origin := range origins {
		for _, part := range strings.Split(origin, ",") {
			part = strings.TrimRight(strings.TrimSpace(part), "/")
			if part != "" && !slices.Contains(cleaned, part) {
				cleaned = append(cleaned, part)
			}
		}
	}

	return cleaned
}

func originAllowed(origin string, allowedOrigins []string) bool {
	origin = strings.TrimRight(strings.TrimSpace(origin), "/")
	if slices.Contains(allowedOrigins, origin) {
		return true
	}

	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "http://localhost" && isLocalhostHTTPOrigin(origin) {
			return true
		}
	}

	return false
}

func isLocalhostHTTPOrigin(origin string) bool {
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}

	return parsed.Scheme == "http" && parsed.Hostname() == "localhost"
}
