package pg

import (
	"context"
	"errors"
	"fmt"

	"shortener/internal/model"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type urlRepository struct {
	db *pgxpool.Pool
}

func NewURLRepository(ctx context.Context, db *pgxpool.Pool) (*urlRepository, error) {
	if err := bootstrap(ctx, db); err != nil {
		return nil, err
	}
	return &urlRepository{db: db}, nil
}

func bootstrap(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx,
		`CREATE TABLE IF NOT EXISTS urls (
		 	uuid SERIAL NOT NULL PRIMARY KEY,
			user_id VARCHAR(50) NOT NULL,
			short_url VARCHAR(10) NOT NULL,
			original_url VARCHAR NOT NULL UNIQUE,
			is_deleted BOOLEAN NOT NULL DEFAULT false
		 )`,
	)
	if err != nil {
		return err
	}

	return nil
}

func (repo *urlRepository) Ping(ctx context.Context) error { return repo.db.Ping(ctx) }

func (repo *urlRepository) Save(ctx context.Context, u model.URLStore) (string, error) {
	_, err := repo.db.Exec(ctx,
		`INSERT INTO urls (user_id, short_url, original_url)
			VALUES ($1, $2, $3);`,
		u.UserID, u.Short, u.Original,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			var shortURL string
			if err := repo.db.QueryRow(ctx,
				"SELECT short_url FROM urls WHERE original_url = $1",
				u.Original,
			).Scan(&shortURL); err != nil {
				return "", fmt.Errorf("pg.Save error: execute select query: %w", err)
			}
			return shortURL, model.ErrURLAlreadyExists
		}
		return "", fmt.Errorf("pg.Save error: execute query: %w", err)
	}

	return u.Short, nil
}

func (repo *urlRepository) SaveAll(ctx context.Context, urls []model.URLStore) error {
	tx, err := repo.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("pg.SaveAll error: start a transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	stmtName := "insert_URL"
	if _, err := tx.Prepare(ctx, stmtName,
		`INSERT INTO urls (user_id, short_url, original_url)
				VALUES ($1, $2, $3);`,
	); err != nil {
		return fmt.Errorf("pg.SaveAll error: create a statement: %w", err)
	}

	batch := &pgx.Batch{}
	for _, u := range urls {
		batch.Queue(stmtName, u.UserID, u.Short, u.Original)
	}

	br := tx.SendBatch(ctx, batch)
	for range urls {
		res, err := br.Exec()
		if err != nil {
			return fmt.Errorf("pg.SaveAll error: batch execute: %w", err)
		}
		if res.RowsAffected() == 0 {
			return fmt.Errorf("pg.SaveAll error: no one row has been saved: %w", err)
		}
	}
	if err := br.Close(); err != nil {
		return fmt.Errorf("pg.SaveAll error: failed to close batch result: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("pg.SaveAll error: failed to commit: %w", err)
	}

	return nil
}

func (repo *urlRepository) Get(ctx context.Context, short string) (model.URLStore, error) {
	var u model.URLStore
	if err := repo.db.QueryRow(ctx,
		`SELECT uuid, short_url, original_url, is_deleted
		FROM urls
		WHERE short_url = $1`,
		short,
	).Scan(&u.UUID, &u.Short, &u.Original, &u.DeletedFlag); err != nil {
		return model.URLStore{}, fmt.Errorf("pg.Get error: failed to find a row: %w", err)
	}

	if u.DeletedFlag {
		return model.URLStore{}, model.ErrDeleted
	}

	return u, nil
}

func (repo *urlRepository) GetByID(ctx context.Context, uuid int) (model.URLStore, error) {
	var u model.URLStore
	if err := repo.db.QueryRow(ctx,
		`SELECT uuid, user_id, short_url, original_url, is_deleted
		FROM urls
		WHERE uuid = $1`,
		uuid,
	).Scan(&u.UUID, &u.UserID, &u.Short, &u.Original, &u.DeletedFlag); err != nil {
		return model.URLStore{}, fmt.Errorf("pg.GetByID error: failed to find a row: %w", err)
	}

	if u.DeletedFlag {
		return model.URLStore{}, model.ErrDeleted
	}

	return u, nil
}

func (repo *urlRepository) GetAllByUser(ctx context.Context, userID string) ([]model.URLStore, error) {
	rows, err := repo.db.Query(ctx,
		`SELECT short_url, original_url
		FROM urls
		WHERE user_id = $1 and is_deleted = false`,
		userID,
	)
	if err != nil {
		return []model.URLStore{}, fmt.Errorf("pg.GetAllByUser error: failed to acquire a collection: %w", err)
	}

	res := make([]model.URLStore, 0)
	for rows.Next() {
		var u model.URLStore
		if err := rows.Scan(&u.Short, &u.Original); err != nil {
			return []model.URLStore{}, fmt.Errorf("pg.GetAllByUser error: failed to scan a row: %w", err)
		}

		res = append(res, u)
	}

	if err := rows.Err(); err != nil {
		return []model.URLStore{}, fmt.Errorf("pg.GetAllByUser error: while reading: %w", err)
	}

	return res, nil
}

func (repo *urlRepository) DeleteBatch(ctx context.Context, userID string, urls []string) error {
	if _, err := repo.db.Exec(ctx,
		`UPDATE urls SET is_deleted=true
		 WHERE user_id = $1 AND short_url = ANY($2);`,
		userID, urls,
	); err != nil {
		return fmt.Errorf("pg.DeleteBatch error: delete: %w", err)
	}

	return nil
}
