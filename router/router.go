// Package router implements routing logic to the corresponding handlers
package router

import (
	"net/http"

	"app/config"
	"app/handlers"
)

func Routes(h *handlers.Handler) *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("GET "+config.RootPrefix, h.Home)

	router.HandleFunc("GET "+config.DashboardPath, h.Dashboard)

	router.HandleFunc("GET "+config.LoginPath, h.Login)

	router.HandleFunc("GET "+config.PricingPath, h.Pricing)

	router.HandleFunc("GET "+config.EditorPath+"{site...}", h.Editor)

	router.HandleFunc("GET /{site}", h.Site)

	return router
}
