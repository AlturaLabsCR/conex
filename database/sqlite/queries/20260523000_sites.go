package queries

import (
	"context"
	"encoding/json"
	"fmt"

	"app/database"
	"app/database/sqlite/db"
)

func (q *SqliteQuerier) CreateSite(ctx context.Context, sub int64, path string, public bool) error {
	return q.queries.CreateSite(ctx, db.CreateSiteParams{
		Sub:    sub,
		Path:   path,
		Public: public,
	})
}

func (q *SqliteQuerier) CreateSiteTimestamps(ctx context.Context, path string) error {
	return q.queries.CreateSiteTimestamps(ctx, path)
}

func (q *SqliteQuerier) CreateSiteClicks(ctx context.Context, path string) error {
	return q.queries.CreateSiteClicks(ctx, path)
}

func (q *SqliteQuerier) SelectSiteByPath(ctx context.Context, path string) (*database.Site, error) {
	site, err := q.queries.SelectSiteByPath(ctx, path)
	if err != nil {
		return nil, err
	}

	return &database.Site{
		Sub:          site.Sub,
		Path:         site.Path,
		Public:       site.Public,
		Name:         site.Name,
		Tags:         databaseTags(site.Tags),
		CreatedAt:    site.CreatedAt,
		LastModified: site.LastModified,
		Clicks:       site.Clicks,
	}, nil
}

func (q *SqliteQuerier) SelectSitesBySub(ctx context.Context, sub int64) ([]database.Site, error) {
	sites, err := q.queries.SelectSitesBySub(ctx, sub)
	if err != nil {
		return nil, err
	}

	out := make([]database.Site, 0, len(sites))
	for _, site := range sites {
		out = append(out, database.Site{
			Sub:          site.Sub,
			Path:         site.Path,
			Public:       site.Public,
			Name:         site.Name,
			Tags:         databaseTags(site.Tags),
			CreatedAt:    site.CreatedAt,
			LastModified: site.LastModified,
			Clicks:       site.Clicks,
		})
	}

	return out, nil
}

func (q *SqliteQuerier) SelectPublicSitesByClicks(ctx context.Context, limit int64, offset int64) ([]database.Site, error) {
	sites, err := q.queries.SelectPublicSitesByClicks(ctx, db.SelectPublicSitesByClicksParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	out := make([]database.Site, 0, len(sites))
	for _, site := range sites {
		out = append(out, database.Site{
			Sub:          site.Sub,
			Path:         site.Path,
			Public:       site.Public,
			Name:         site.Name,
			Tags:         databaseTags(site.Tags),
			CreatedAt:    site.CreatedAt,
			LastModified: site.LastModified,
			Clicks:       site.Clicks,
		})
	}

	return out, nil
}

func (q *SqliteQuerier) SelectPublicSitesByCreatedAt(ctx context.Context, limit int64, offset int64) ([]database.Site, error) {
	sites, err := q.queries.SelectPublicSitesByCreatedAt(ctx, db.SelectPublicSitesByCreatedAtParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	out := make([]database.Site, 0, len(sites))
	for _, site := range sites {
		out = append(out, database.Site{
			Sub:          site.Sub,
			Path:         site.Path,
			Public:       site.Public,
			Name:         site.Name,
			Tags:         databaseTags(site.Tags),
			CreatedAt:    site.CreatedAt,
			LastModified: site.LastModified,
			Clicks:       site.Clicks,
		})
	}

	return out, nil
}

func (q *SqliteQuerier) SearchPublicSites(ctx context.Context, query string, limit int64, offset int64) ([]database.Site, error) {
	sites, err := q.queries.SearchPublicSites(ctx, db.SearchPublicSitesParams{
		Column1: query,
		Column2: limit,
		Column3: offset,
	})
	if err != nil {
		return nil, err
	}

	out := make([]database.Site, 0, len(sites))
	for _, site := range sites {
		out = append(out, database.Site{
			Sub:          site.Sub,
			Path:         site.Path,
			Public:       site.Public,
			Name:         site.Name,
			Tags:         databaseTags(site.Tags),
			CreatedAt:    site.CreatedAt,
			LastModified: site.LastModified,
			Clicks:       site.Clicks,
		})
	}

	return out, nil
}

func (q *SqliteQuerier) UpsertSiteName(ctx context.Context, path string, name string) error {
	return q.queries.UpsertSiteName(ctx, db.UpsertSiteNameParams{
		Path: path,
		Name: name,
	})
}

func (q *SqliteQuerier) DeleteSiteTags(ctx context.Context, path string) error {
	return q.queries.DeleteSiteTags(ctx, path)
}

func (q *SqliteQuerier) CreateSiteTag(ctx context.Context, path string, tag string) error {
	return q.queries.CreateSiteTag(ctx, db.CreateSiteTagParams{
		Path: path,
		Tag:  tag,
	})
}

func (q *SqliteQuerier) UpdateSitePublic(ctx context.Context, sub int64, path string, public bool) error {
	return q.queries.UpdateSitePublic(ctx, db.UpdateSitePublicParams{
		Public: public,
		Sub:    sub,
		Path:   path,
	})
}

func (q *SqliteQuerier) IncrementSiteClicks(ctx context.Context, path string) error {
	return q.queries.IncrementSiteClicks(ctx, path)
}

func (q *SqliteQuerier) UpdateSiteLastModified(ctx context.Context, path string) error {
	return q.queries.UpdateSiteLastModified(ctx, path)
}

func (q *SqliteQuerier) DeleteSite(ctx context.Context, sub int64, path string) (*database.Site, error) {
	site, err := q.queries.DeleteSite(ctx, db.DeleteSiteParams{
		Sub:  sub,
		Path: path,
	})
	if err != nil {
		return nil, err
	}

	return &database.Site{
		Sub:          site.Sub,
		Path:         site.Path,
		Public:       site.Public,
		Name:         site.Path_2,
		Tags:         databaseTags(site.Column5),
		CreatedAt:    site.Column6,
		LastModified: site.Column7,
		Clicks:       site.Column8,
	}, nil
}

func databaseTags(tags any) []string {
	switch tags := tags.(type) {
	case nil:
		return nil
	case []string:
		return tags
	case []any:
		out := make([]string, 0, len(tags))
		for _, tag := range tags {
			out = append(out, fmt.Sprint(tag))
		}
		return out
	case string:
		var out []string
		if err := json.Unmarshal([]byte(tags), &out); err == nil {
			return out
		}
		return []string{tags}
	case []byte:
		var out []string
		if err := json.Unmarshal(tags, &out); err == nil {
			return out
		}
		return []string{string(tags)}
	default:
		return []string{fmt.Sprint(tags)}
	}
}
