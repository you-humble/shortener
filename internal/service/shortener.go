package service

import (
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"
)

type URLRepository interface {
	Save(string, string) error
	Get(string) (string, error)
}

type urlService struct {
	baseAddr string
	repo     URLRepository
}

func NewURLService(baseAddr string, repo URLRepository) *urlService {
	return &urlService{baseAddr: strings.TrimRight(baseAddr, "/"), repo: repo}
}

func (s *urlService) GenerateShortURL(scheme, original string) (string, error) {
	if original == "" {
		return "", errors.New("original URL is empty")
	}
	if scheme == "" {
		return "", errors.New("scheme is empty")
	}

	short := genShortURL(original)

	if err := s.repo.Save(short, original); err != nil {
		return "", err
	}

	if hasScheme(s.baseAddr) {
		return fmt.Sprintf("%s/%s", s.baseAddr, short), nil
	}
	return fmt.Sprintf("%s://%s/%s", scheme, s.baseAddr, short), nil
}

func (s *urlService) OriginalURL(short string) (string, error) {
	if short == "" {
		return "", fmt.Errorf("empty path")
	}

	original, err := s.repo.Get(short)
	if err != nil {
		return "", err
	}

	return original, nil
}

func genShortURL(url string) string {
	hash := sha256.Sum256([]byte(url))
	base := base32.HexEncoding.EncodeToString(hash[:])
	result := strings.ReplaceAll(base[:8], "=", "")

	return result
}

func hasScheme(addr string) bool {
    return strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://")
}