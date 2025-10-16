// Package handlers implements rendering functions for endpoints
package handlers

import (
	"database/sql"
	"log/slog"
)

type Handler struct {
	params HandlerParams
}

type HandlerParams struct {
	Production bool
	DB         *sql.DB
	Logger     *slog.Logger
}

func New(p HandlerParams) *Handler {
	return &Handler{
		params: p,
	}
}

func (h *Handler) Prod() bool {
	return h.params.Production
}

func (h *Handler) DB() *sql.DB {
	return h.params.DB
}

func (h *Handler) Log() *slog.Logger {
	return h.params.Logger
}
