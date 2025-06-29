package repo

import (
	"sync"
	
	"shortener/internal/model"
)

type urlRepository struct {
	mu sync.Mutex
	db map[string]string
}

func NewURLRepository() *urlRepository {
	return &urlRepository{mu: sync.Mutex{}, db: make(map[string]string, 100)}
}

func (repo *urlRepository) Save(key, val string) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	repo.db[key] = val

	return nil
}

func (repo *urlRepository) Get(key string) (string, error) {
	repo.mu.Lock()
	val, ok := repo.db[key]
	repo.mu.Unlock()
	if !ok {
		return "", model.ErrURLNotFound
	}

	return val, nil
}
