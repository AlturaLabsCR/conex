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

	router.HandleFunc("GET "+config.RegisterPath, h.RegisterForm)
	router.HandleFunc("POST "+config.RegisterPath, h.Register)
	router.HandleFunc("PUT "+config.RegisterPath, h.RegisterConfirm)

	router.HandleFunc("GET "+config.LoginPath, h.LoginForm)
	router.HandleFunc("POST "+config.LoginPath, h.Login)
	router.HandleFunc("PUT "+config.LoginPath, h.LoginConfirm)

	router.HandleFunc("GET "+config.DashboardPath, h.Dashboard)
	router.HandleFunc("GET "+config.PricingPath, h.Pricing)
	router.HandleFunc("GET "+config.EditorPath+"{site...}", h.Editor)

	router.HandleFunc("GET "+config.RootPrefix+"{site}", h.Site)

	return router
}
