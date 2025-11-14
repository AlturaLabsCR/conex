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

	router.HandleFunc("GET "+config.Endpoints[config.RootPath], h.Home)

	router.HandleFunc("GET "+config.Endpoints[config.RegisterPath], h.RegisterForm)
	router.HandleFunc("PUT "+config.Endpoints[config.RegisterPath], h.Register)
	router.HandleFunc("POST "+config.Endpoints[config.RegisterPath], h.RegisterConfirm)

	router.HandleFunc("GET "+config.Endpoints[config.LoginPath], h.LoginForm)
	router.HandleFunc("PUT "+config.Endpoints[config.LoginPath], h.Login)
	router.HandleFunc("POST "+config.Endpoints[config.LoginPath], h.LoginConfirm)

	router.HandleFunc("GET "+config.Endpoints[config.PricingPath], h.Pricing)

	loggedIn := middleware.Stack(
		h.AuthenticationMiddleware(
			false,
			0,
			config.Endpoints[config.LoginPath],
		),
	)

	router.Handle("GET "+config.Endpoints[config.EditorPath]+"{site...}", middleware.With(loggedIn, h.Editor))
	router.Handle("GET "+config.Endpoints[config.DashboardPath], middleware.With(loggedIn, h.Dashboard))
	router.Handle("GET "+config.Endpoints[config.AccountPath], middleware.With(loggedIn, h.Account))
	router.Handle("GET "+config.Endpoints[config.LogoutPath], middleware.With(loggedIn, h.Logout))

	protected := middleware.Stack(
		h.AuthenticationMiddleware(
			true,
			0,
			"",
		),
	)

	router.Handle("DELETE "+config.Endpoints[config.LogoutPath]+"/{sessionID}", middleware.With(protected, h.LogoutAskedSession))

	router.Handle("POST "+config.Endpoints[config.EditorPath], middleware.With(protected, h.NewSite))
	router.Handle("PUT "+config.Endpoints[config.EditorPath], middleware.With(protected, h.Publish))

	router.Handle("POST "+config.Endpoints[config.UploadPath], middleware.With(protected, h.Upload))

	router.HandleFunc("GET "+config.Endpoints[config.RootPath]+"{site}", h.Site)

	router.Handle("PATCH "+config.Endpoints[config.SettingsPath], middleware.With(protected, h.UpdateSettings))

	return router
}
