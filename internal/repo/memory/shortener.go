package memory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"shortener/internal/model"
)

var nextUUID int = 1

type urlRepository struct {
	mu sync.Mutex
	db map[string][]byte
}

func NewURLRepository() (*urlRepository, error) {
	return &urlRepository{mu: sync.Mutex{}, db: make(map[string][]byte, 100)}, nil
}

func (repo *urlRepository) Save(ctx context.Context, u model.URLStore) (string, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	u.UUID = nextUUID
	b, err := json.Marshal(u)
	if err != nil {
		return "", fmt.Errorf("memory.Save error: marshal error: %w", err)
	}

	repo.db[u.Short] = b
	nextUUID++

	return u.Short, nil
}

func (repo *urlRepository) Get(ctx context.Context, key string) (model.URLStore, error) {
	repo.mu.Lock()
	val, ok := repo.db[key]
	repo.mu.Unlock()
	if !ok {
		return model.URLStore{}, model.ErrURLNotFound
	}

	var u model.URLStore
	if err := json.Unmarshal(val, &u); err != nil {
		return model.URLStore{}, fmt.Errorf("memory.Save error: unmarshal error: %w", err)
	}

	return u, nil
}

func (repo *urlRepository) Ping(ctx context.Context) error {
	return errors.New("unimplemented")
}

func (repo *urlRepository) SaveAll(ctx context.Context, urls []model.URLStore) error {
	return errors.New("unimplemented")
}

func (repo *urlRepository) GetAllByUser(ctx context.Context, userID string) ([]model.URLStore, error) {
	return []model.URLStore{}, errors.New("unimplemented")
}

func (repo *urlRepository) DeleteBatch(ctx context.Context, userID string, urls []string) error {
	return errors.New("unimplemented")
}

func (repo *urlRepository) GetByID(ctx context.Context, uuid int) (model.URLStore, error) {
	return model.URLStore{}, errors.New("unimplemented")
}
