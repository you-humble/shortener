package file

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"shortener/internal/model"
)

var nextUUID int = 1

type urlRepository struct {
	mu sync.Mutex
	db string
}

func NewURLRepository(filePath string) (*urlRepository, error) {
	return &urlRepository{mu: sync.Mutex{}, db: filePath}, nil
}

func (repo *urlRepository) Save(ctx context.Context, u model.URLStore) (string, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	u.UUID = nextUUID

	f, err := os.OpenFile(repo.db, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return "", fmt.Errorf("file.Save error: open file: %w", err)
	}

	if err := json.NewEncoder(f).Encode(u); err != nil {
		return "", fmt.Errorf("file.Save error: marshal error: %w", err)
	}
	nextUUID++

	return u.Short, nil
}

func (repo *urlRepository) Get(ctx context.Context, short string) (model.URLStore, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	f, err := os.OpenFile(repo.db, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return model.URLStore{}, fmt.Errorf("file.Get error: open file error: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if bytes.Contains(line, []byte(short)) {
			var u model.URLStore
			if err := json.Unmarshal(line, &u); err != nil {
				return model.URLStore{}, fmt.Errorf("file.Get error: unmarshal error: %w", err)
			}
			return u, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return model.URLStore{}, fmt.Errorf("file.Get error: scanner error: %w", err)
	}

	return model.URLStore{}, model.ErrURLNotFound
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
