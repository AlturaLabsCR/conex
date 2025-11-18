// Package handlers implements rendering functions for endpoints
package handlers

import (
	"compress/gzip"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"

	"app/i18n"
	"app/internal/db"
	"app/sessions"
	"app/utils/smtp"
)

type Handler struct {
	params     HandlerParams
	Translator func(*http.Request) func(string) string
	Sessions   *sessions.Store[db.Session]
}

type HandlerParams struct {
	Production   bool
	Logger       *slog.Logger
	Queries      *db.Queries
	Pool         *pgxpool.Pool
	Storage      *s3.Client
	Locales      map[string]map[string]string
	SMTPAuth     smtp.AuthParams
	CookieName   string
	CookiePath   string
	ServerSecret string
}

type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")

		gz := gzip.NewWriter(w)
		defer gz.Close()

		gzw := gzipResponseWriter{
			ResponseWriter: w,
			Writer:         gz,
		}

		next.ServeHTTP(gzw, r)
	})
}

func New(params HandlerParams) *Handler {
	sessions := sessions.New[db.Session](sessions.StoreParams{
		CookieName:     params.CookieName,
		CookiePath:     params.CookiePath,
		CookieSameSite: http.SameSiteStrictMode,
		CookieTTL:      24 * time.Hour,
		JWTSecret:      params.ServerSecret,
	})

	translator := i18n.New(params.Locales).TranslateHTTPRequest

	return &Handler{
		params:     params,
		Translator: translator,
		Sessions:   sessions,
	}
}

func (h *Handler) Prod() bool {
	return h.params.Production
}

func (h *Handler) DB() *pgxpool.Pool {
	return h.params.Pool
}

func (h *Handler) Queries() *db.Queries {
	return h.params.Queries
}

func (h *Handler) S3() *s3.Client {
	return h.params.Storage
}

func (h *Handler) Log() *slog.Logger {
	return h.params.Logger
}

func (h *Handler) SMTPClient() *smtp.Auth {
	return smtp.Client(h.params.SMTPAuth)
}
