package handlers

import (
	"errors"
	"net/http"

	"app/internal/db"
)

const maxMemory int64 = 10 << 20

func (h *Handler) UploadObject(w http.ResponseWriter, r *http.Request, inputName string, queries *db.Queries) (db.SiteObject, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxMemory)

	siteSlug := r.PathValue("site")
	if siteSlug == "" {
		return db.SiteObject{}, errors.New("missing site pathvalue")
	}

	if err := r.ParseMultipartForm(maxMemory); err != nil {
		return db.SiteObject{}, err
	}

	file, header, err := r.FormFile(inputName)
	if err != nil {
		return db.SiteObject{}, err
	}
	defer file.Close()

	obj, err := h.PutObject(r.Context(), siteSlug, header.Filename, file, queries)
	if err != nil {
		return db.SiteObject{}, err
	}

	return obj, nil
}
