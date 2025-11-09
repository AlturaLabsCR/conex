// Package router implements routing logic to the corresponding handlers
package router

import (
	"net/http"

	"app/config"
	"app/handlers"
	"app/middleware"
)

func Routes(h *handlers.Handler) *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("GET "+config.RootPrefix, h.Home)

	router.HandleFunc("GET "+config.RegisterPath, h.RegisterForm)
	router.HandleFunc("PUT "+config.RegisterPath, h.Register)
	router.HandleFunc("POST "+config.RegisterPath, h.RegisterConfirm)

	router.HandleFunc("GET "+config.LoginPath, h.LoginForm)
	router.HandleFunc("PUT "+config.LoginPath, h.Login)
	router.HandleFunc("POST "+config.LoginPath, h.LoginConfirm)

	router.HandleFunc("GET "+config.PricingPath, h.Pricing)

	loggedIn := middleware.Stack(
		h.AuthenticationMiddleware(
			false,
			0,
			config.LoginPath,
		),
	)

	router.Handle("GET "+config.EditorPath+"{site...}", middleware.With(loggedIn, h.Editor))
	router.Handle("GET "+config.DashboardPath, middleware.With(loggedIn, h.Dashboard))
	router.Handle("GET "+config.AccountPath, middleware.With(loggedIn, h.Account))
	router.Handle("GET "+config.LogoutPath, middleware.With(loggedIn, h.Logout))

	protected := middleware.Stack(
		h.AuthenticationMiddleware(
			true,
			0,
			"",
		),
	)

	router.Handle("DELETE "+config.LogoutPath+"/{sessionID}", middleware.With(protected, h.LogoutAskedSession))

	router.Handle("POST "+config.EditorPath, middleware.With(protected, h.NewSite))

	router.HandleFunc("GET "+config.RootPrefix+"{site}", h.Site)

	return router
}
