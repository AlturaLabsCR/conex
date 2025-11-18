package config

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDB(ctx context.Context) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dbConn)
	if err != nil {
		return nil, err
	}
	return pool, nil
}
