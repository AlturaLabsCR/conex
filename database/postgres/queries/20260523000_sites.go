package queries

import (
	"context"

	"app/database"
	"app/database/postgres/db"
)

func (q *PostgresQuerier) CreateSite(ctx context.Context, sub int64, path string, public bool) error {
	return q.queries.CreateSite(ctx, db.CreateSiteParams{
		Sub:    sub,
		Path:   path,
		Public: public,
	})
}

func (q *PostgresQuerier) SelectSiteByPath(ctx context.Context, path string) (*database.Site, error) {
	site, err := q.queries.SelectSiteByPath(ctx, path)
	if err != nil {
		return nil, err
	}

	return &database.Site{
		Sub:    site.Sub,
		Path:   site.Path,
		Public: site.Public,
	}, nil
}

func (q *PostgresQuerier) SelectSitesBySub(ctx context.Context, sub int64) ([]database.Site, error) {
	sites, err := q.queries.SelectSitesBySub(ctx, sub)
	if err != nil {
		return nil, err
	}

	out := make([]database.Site, 0, len(sites))
	for _, site := range sites {
		out = append(out, database.Site{
			Sub:    site.Sub,
			Path:   site.Path,
			Public: site.Public,
		})
	}

	return out, nil
}

func (q *PostgresQuerier) UpdateSitePublic(ctx context.Context, sub int64, path string, public bool) (*database.Site, error) {
	site, err := q.queries.UpdateSitePublic(ctx, db.UpdateSitePublicParams{
		Sub:    sub,
		Path:   path,
		Public: public,
	})
	if err != nil {
		return nil, err
	}

	return &database.Site{
		Sub:    site.Sub,
		Path:   site.Path,
		Public: site.Public,
	}, nil
}

func (q *PostgresQuerier) DeleteSite(ctx context.Context, sub int64, path string) (*database.Site, error) {
	site, err := q.queries.DeleteSite(ctx, db.DeleteSiteParams{
		Sub:  sub,
		Path: path,
	})
	if err != nil {
		return nil, err
	}

	return &database.Site{
		Sub:    site.Sub,
		Path:   site.Path,
		Public: site.Public,
	}, nil
}
