package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"app/internal/db"
	"app/templates"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	queries := db.New(h.DB())

	sites, err := queries.GetHomePageSitesWithMetricsFromMostTotalVisits(ctx)
	if err != nil {
		h.Log().Error("error querying sites with metrics", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tr := h.Translator(r)

	header := templates.HomeHeader(tr)
	content := templates.CardsGrid(tr, sites, false)

	if err := templates.Base(tr, header, content, true).Render(ctx, w); err != nil {
		h.Log().Error("error rendering template", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	tr := h.Translator(r)
	ctx := r.Context()

	jsonStr := r.URL.Query().Get("datastar")

	var payload struct {
		Query string `json:"search"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	queries := db.New(h.DB())

	sites, err := queries.GetPublishedSitesWithMetrics(ctx)
	if err != nil {
		h.Log().Error("error querying sites with metrics", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	qq := strings.Fields(
		strings.ReplaceAll(
			strings.ToLower(strings.TrimSpace(payload.Query)),
			",",
			" ",
		),
	)
	if len(qq) == 0 {
		if err := templates.CardsGrid(tr, sites, false).Render(ctx, w); err != nil {
			h.Log().Error("error rendering template", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}

	matched := make([]db.SitesWithMetric, 0, len(sites))

	for _, s := range sites {
		title := strings.ToLower(s.SiteTitle)
		desc := strings.ToLower(s.SiteDescription)

		type Tag struct {
			Name  string `json:"name"`
			Color string `json:"color"`
		}

		var tagsArr []Tag
		_ = json.Unmarshal([]byte(s.SiteTagsJson), &tagsArr)

		var tagNames []string
		for _, t := range tagsArr {
			tagNames = append(tagNames, t.Name)
		}

		tags := strings.ToLower(strings.Join(tagNames, " "))

		siteStr := title + " " + desc + " " + tags

		for _, q := range qq {
			if fuzzy.MatchNormalized(q, siteStr) {
				matched = append(matched, s)
				break
			}
		}
	}

	h.Log().Debug("sites with metric matched", "count", len(matched), "sites", matched)

	if err := templates.CardsGrid(tr, matched, false).Render(ctx, w); err != nil {
		h.Log().Error("error rendering template", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
