package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"shortener/internal/model"
	"shortener/internal/shared/logger"
	"strings"
)

type URLService interface {
	GenerateShortURL(string, string) (string, error)
	OriginalURL(string) (string, error)
}

type urlHandler struct {
	log *logger.Logger
	svc URLService
}

func NewURLHandler(log *logger.Logger, svc URLService) *urlHandler {
	return &urlHandler{log: log, svc: svc}
}

func (h *urlHandler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	shortURL := strings.Trim(r.URL.Path, "/")
	originalURL, err := h.svc.OriginalURL(shortURL)
	if err != nil {
		h.log.Error("RedirectURL", logger.Error(err))
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func (h *urlHandler) ShortenURLText(w http.ResponseWriter, r *http.Request) {
	b, err := readBody(r)
	if err != nil {
		h.log.Error("ShortenURLText", logger.Error(err))
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	originalURL := strings.TrimSpace(string(b))

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	resp, err := h.svc.GenerateShortURL(scheme, originalURL)
	if err != nil {
		h.log.Error("ShortenURLText", logger.Error(err))
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resp))
}
func (h *urlHandler) ShortenURLJSON(w http.ResponseWriter, r *http.Request) {
	b, err := readBody(r)
	if err != nil {
		h.log.Error("ShortenURLJSON", logger.Error(err))
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var urlRecv model.ShortenRequest
	if err := json.Unmarshal(b, &urlRecv); err != nil {
		h.log.Error("ShortenURLJSON", logger.Error(err))
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if urlRecv.URL == "" {
		h.log.Error("ShortenURLJSON", logger.ErrorS("empty URL"))
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	resp, err := h.svc.GenerateShortURL(scheme, urlRecv.URL)
	if err != nil {
		h.log.Error("ShortenURLJSON", logger.Error(err))
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(model.ShortenResponse{Result: resp}); err != nil {
		h.log.Error("ShortenURLJSON", logger.Error(err))
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}

func readBody(r *http.Request) ([]byte, error) {
	b, err := io.ReadAll(r.Body)
	if err != nil || len(b) == 0 {
		return nil, err
	}
	defer r.Body.Close()

	if len(b) == 0 {
		return nil, errors.New("request body is empty")
	}

	return b, nil
}
