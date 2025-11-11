package handlers

import (
	"encoding/json"
	"net/http"
)

const maxMemory int64 = 10 << 20

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxMemory)

	if err := r.ParseMultipartForm(maxMemory); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// TODO: S3 logic
	// - Limit size
	// - User has plan that allows image uploading
	// - Check for user's objects and verify it doesnt exceed quota
	fileURL := "https://pet-health-content-media.chewy.com/wp-content/uploads/2024/09/11170344/202106american-eskimo-dog-5.jpg"

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": 1,
		"file": map[string]any{
			"url": fileURL,
		},
	})
}
