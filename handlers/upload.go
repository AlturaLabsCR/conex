package handlers

import (
	"encoding/json"
	"net/http"
)

const maxMemory int64 = 10 << 20

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxMemory)

	siteSlug := r.PathValue("site")
	if siteSlug == "" {
		w.WriteHeader(http.StatusBadRequest)
		h.Log().Debug("missing site pathvlaue")
		return
	}

	if err := r.ParseMultipartForm(maxMemory); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.Log().Debug("file too large", "error", err)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.Log().Debug("error retrieving file", "error", err)
		return
	}
	defer file.Close()

	fileURL, err := h.PutObject(r.Context(), siteSlug, header.Filename, file)
	if err != nil {
		http.Error(w, "Upload failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.Log().Debug("uploaded file successfully", "url", fileURL)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": 1,
		"file": map[string]any{
			"url": fileURL,
		},
	})
}
