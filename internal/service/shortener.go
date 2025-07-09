package service

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"

	"shortener/internal/model"
)

type URLRepository interface {
	Ping(context.Context) error
	Save(context.Context, model.URLStore) (string, error)
	SaveAll(context.Context, []model.URLStore) error
	Get(context.Context, string) (model.URLStore, error)
	GetAllByUser(context.Context, string) ([]model.URLStore, error)
}

type urlService struct {
	baseAddr string
	repo     URLRepository
}

func NewURLService(baseAddr string, repo URLRepository) *urlService {
	return &urlService{baseAddr: strings.TrimRight(baseAddr, "/"), repo: repo}
}

func (s *urlService) GenerateShortURL(ctx context.Context, scheme, userID, original string) (string, error) {
	if original == "" {
		return "", errors.New("original URL is empty")
	}
	if scheme == "" {
		return "", errors.New("scheme is empty")
	}

	shortURL, err := s.repo.Save(ctx,
		model.URLStore{
			UserID:   userID,
			Short:    genShortURL(original),
			Original: original},
	)
	if err != nil {
		if errors.Is(err, model.ErrURLAlreadyExists) {
			return s.shortWithScheme(scheme, shortURL), model.ErrURLAlreadyExists
		}
		return "", err
	}

	return s.shortWithScheme(scheme, shortURL), nil
}

func (s *urlService) GenerateShortBatch(
	ctx context.Context,
	scheme string,
	userID string,
	req []model.ShortenBatchRequest,
) ([]model.ShortenBatchResponse, error) {
	if scheme == "" {
		return []model.ShortenBatchResponse{}, errors.New("scheme is empty")
	}

	urls := make([]model.URLStore, 0, len(req))
	for _, u := range req {
		urls = append(urls, model.URLStore{
			UserID:   userID,
			Short:    genShortURL(u.Original),
			Original: u.Original,
		})
	}

	if err := s.repo.SaveAll(ctx, urls); err != nil {
		return []model.ShortenBatchResponse{}, err
	}

	res := make([]model.ShortenBatchResponse, len(req))
	for i := range req {
		res[i].Short = s.shortWithScheme(scheme, urls[i].Short)
		res[i].CorrelationID = req[i].CorrelationID
	}

	return res, nil
}

func (s *urlService) OriginalURL(ctx context.Context, short string) (string, error) {
	if short == "" {
		return "", errors.New("empty path")
	}

	u, err := s.repo.Get(ctx, short)
	if err != nil {
		return "", err
	}

	return u.Original, nil
}

func (s *urlService) UserStore(ctx context.Context, userID string) ([]model.URLStore, error) {
	res, err := s.repo.GetAllByUser(ctx, userID)
	if err != nil {
		return []model.URLStore{}, err
	}

	for i := range res {
		res[i].Short = s.shortWithScheme("http", res[i].Short)

	}

	return res, nil
}

func (s *urlService) Ping(ctx context.Context) error { return s.repo.Ping(ctx) }

func genShortURL(url string) string {
	hash := sha256.Sum256([]byte(url))
	base := base32.HexEncoding.EncodeToString(hash[:])
	result := strings.ReplaceAll(base[:8], "=", "")

	return result
}

func hasScheme(addr string) bool {
	return strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://")
}

func (s *urlService) shortWithScheme(scheme, shortURL string) string {
	if hasScheme(s.baseAddr) {
		return fmt.Sprintf("%s/%s", s.baseAddr, shortURL)
	}
	return fmt.Sprintf("%s://%s/%s", scheme, s.baseAddr, shortURL)
}
