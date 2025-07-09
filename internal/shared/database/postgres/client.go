package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)



func NewConnect(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("postgres.NewConnect - unable to create new pg pool: %w", err)
	}

	db.Config().MaxConns = 10

	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("postgres.NewConnect - failed to connect to postgres server: %w", err)
	}

	return db, nil
}

func Ping(ctx context.Context, db *pgxpool.Pool) error { return db.Ping(ctx) }
