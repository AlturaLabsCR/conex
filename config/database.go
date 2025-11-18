package config

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func InitDB(ctx context.Context) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, dbConn)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
