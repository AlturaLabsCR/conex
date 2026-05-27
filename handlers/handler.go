// Package handlers registers the application's HTTP handlers.
package handlers

import (
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	appauth "app/auth"
	"app/database"
	"app/mailer"
	"app/middleware"
	"app/sites"
	"github.com/tavocg/go-auth"
	"github.com/tavocg/go-i18n"
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type Handler struct {
	initialized bool
	dev         bool

	logger        Logger
	db            database.Database
	mailer        *mailer.Mailer
	sites         sites.Sites
	authenticator auth.Authenticator[*appauth.Claims]
	localizer     *i18n.Localizer
	rootPrefix    string
	baseURL       string

	mux   *http.ServeMux
	paths []string
}

type Options struct {
	Logger        Logger
	Dev           bool
	DB            database.Database
	Mailer        *mailer.Mailer
	Sites         sites.Sites
	Authenticator auth.Authenticator[*appauth.Claims]
	Localizer     *i18n.Localizer
	RootPrefix    string
	BaseURL       string
}

func NewHandler(opts Options) *Handler {
	logger := opts.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	localizer := opts.Localizer
	if localizer == nil {
		panic("handler localizer is required")
	}

	authenticator := opts.Authenticator
	if authenticator == nil {
		panic("handler authenticator is required")
	}

	if opts.Mailer == nil {
		panic("handler mailer is required")
	}
	if opts.Sites == nil {
		panic("handler sites service is required")
	}
	if opts.DB == nil {
		panic("handler database is required")
	}

	rootPrefix := normalizeRootPrefix(opts.RootPrefix)
	baseURL := normalizeBaseURL(opts.BaseURL)

	next := &Handler{
		initialized:   true,
		dev:           opts.Dev,
		logger:        logger,
		db:            opts.DB,
		mailer:        opts.Mailer,
		sites:         opts.Sites,
		authenticator: authenticator,
		localizer:     localizer,
		rootPrefix:    rootPrefix,
		baseURL:       baseURL,
		mux:           http.NewServeMux(),
	}

	next.registerRoutes()

	return next
}

func (h *Handler) Add(method, path string, fn http.HandlerFunc) {
	h.AddHandler(method, path, fn)
}

func (h *Handler) AddHandler(method, path string, handler http.Handler) {
	pattern := path
	if method != "" {
		pattern = method + " " + path
	}

	h.paths = append(h.paths, path)
	h.mux.Handle(pattern, middleware.RequestLogger(h.logger, pattern, handler))
}

func (h *Handler) Mux() *http.ServeMux {
	if h == nil || !h.initialized {
		panic("handler not initialized")
	}

	return h.mux
}

func (h *Handler) registerRoutes() {
	h.registerAuthRoutes()
	h.registerAccountRoutes()
	h.registerSiteRoutes()
	h.registerRootRoutes()
	h.registerStaticRoutes()
}

func (h *Handler) routePath(route string) string {
	route = strings.TrimSpace(route)
	if route == "" || route == "/" {
		if h.rootPrefix == "" {
			return "/"
		}

		return h.rootPrefix
	}

	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}

	return h.rootPrefix + route
}

func (h *Handler) siteURL(path string) string {
	route := h.routePath("/" + path)
	if h.baseURL == "" {
		return route
	}

	return strings.TrimRight(h.baseURL, "/") + route
}

func normalizeRootPrefix(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" || prefix == "/" {
		return ""
	}

	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}

	return strings.TrimRight(prefix, "/")
}

func normalizeBaseURL(baseURL string) string {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return ""
	}

	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return strings.TrimRight(baseURL, "/")
	}

	return strings.TrimRight(parsed.String(), "/")
}
