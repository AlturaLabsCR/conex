// Package router implements routing logic to the corresponding handlers
package router

import (
	"net/http"

	"app/handlers"
)

func Routes(h *handlers.Handler) *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("GET /", h.Home)

	router.HandleFunc("GET /d", h.Dashboard)

	router.HandleFunc("GET /login", h.Login)

	router.HandleFunc("GET /{site}", h.Site)

	router.HandleFunc("GET /e/{site...}", h.Editor)

	return router
}
