package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"shortener/internal/model"
	"shortener/internal/shared/logger"
)

type URLService interface {
	Ping(context.Context) error
	GenerateShortURL(context.Context, string, string, string) (string, error)
	GenerateShortBatch(context.Context, string, string, []model.ShortenBatchRequest) ([]model.ShortenBatchResponse, error)
	OriginalURL(context.Context, string) (string, error)
	URLByID(context.Context, int) (model.URLStore, error)
	UserStore(context.Context, string) ([]model.URLStore, error)
	MakeDeleted(context.Context, model.DeleteURLsRequest)
}

type AuthService interface {
	UserIDFromContext(context.Context) (string, bool)
}

type urlHandler struct {
	log  *logger.Logger
	svc  URLService
	auth AuthService
}

func NewURLHandler(log *logger.Logger, svc URLService, auth AuthService) *urlHandler {
	return &urlHandler{log: log, svc: svc, auth: auth}
}

func (h *urlHandler) URLByID(w http.ResponseWriter, r *http.Request) {
	urlID, err := strconv.Atoi(strings.Trim(r.URL.Path, "/"))
	if err != nil {
		h.log.Error("URLByID", logger.Error(err))
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	res, err := h.svc.URLByID(r.Context(), urlID)
	if err != nil {
		if errors.Is(err, model.ErrDeleted) {
			w.WriteHeader(http.StatusGone)
			return
		}
		h.log.Error("URLByID", logger.Error(err))
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		h.log.Error("URLByID", logger.Error(err))
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}

func (h *urlHandler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	shortURL := strings.Trim(r.URL.Path, "/")
	originalURL, err := h.svc.OriginalURL(r.Context(), shortURL)
	if err != nil {
		if errors.Is(err, model.ErrDeleted) {
			w.WriteHeader(http.StatusGone)
			return
		}
		h.log.Error("RedirectURL", logger.Error(err))
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func (h *urlHandler) AllUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.auth.UserIDFromContext(r.Context())
	if !ok {
		h.log.Error("AllUserURLs", logger.ErrorS("unauthorized user"))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	resp, err := h.svc.UserStore(r.Context(), userID)
	if err != nil {
		h.log.Error("AllUserURLs", logger.Error(err))
		http.NotFound(w, r)
		return
	}
	if len(resp) == 0 {
		h.log.Info("AllUserURLs", logger.ErrorS("empty user store"))
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.log.Error("AllUserURLs", logger.Error(err))
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}

func (h *urlHandler) ShortenURLText(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.auth.UserIDFromContext(r.Context())
	if !ok {
		h.log.Error("AllUserURLs", logger.ErrorS("unauthorized user"))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	b, err := readBody(r)
	if err != nil || len(b) == 0 {
		h.log.Error("ShortenURLText", logger.Error(err))
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	originalURL := strings.TrimSpace(string(b))

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	resp, err := h.svc.GenerateShortURL(r.Context(), scheme, userID, originalURL)
	if err != nil {
		if errors.Is(err, model.ErrURLAlreadyExists) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(resp))
			return
		}
		h.log.Error("ShortenURLText", logger.Error(err))
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resp))
}
func (h *urlHandler) ShortenURLJSON(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.auth.UserIDFromContext(r.Context())
	if !ok {
		h.log.Error("AllUserURLs", logger.ErrorS("unauthorized user"))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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

	resp, err := h.svc.GenerateShortURL(r.Context(), scheme, userID, urlRecv.URL)
	if err != nil {
		if errors.Is(err, model.ErrURLAlreadyExists) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			if err := json.NewEncoder(w).Encode(model.ShortenResponse{Result: resp}); err != nil {
				h.log.Error("ShortenURLJSON", logger.Error(err))
				http.Error(w, "Internal error", http.StatusInternalServerError)
				return
			}
			return
		}
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

func (h *urlHandler) ShortenBatchJSON(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.auth.UserIDFromContext(r.Context())
	if !ok {
		h.log.Error("AllUserURLs", logger.ErrorS("unauthorized user"))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	var urlRecv []model.ShortenBatchRequest
	if err := dec.Decode(&urlRecv); err != nil {
		h.log.Error("ShortenBatchJSON", logger.Error(err))
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		h.log.Error("ShortenBatchJSON", logger.ErrorS("empty URL"))
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	resp, err := h.svc.GenerateShortBatch(r.Context(), scheme, userID, urlRecv)
	if err != nil {
		h.log.Error("ShortenBatchJSON", logger.Error(err))
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.log.Error("ShortenBatchJSON", logger.Error(err))
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}

func (h *urlHandler) DeleteURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.auth.UserIDFromContext(r.Context())
	if !ok {
		h.log.Error("DeleteURLs", logger.ErrorS("unauthorized user"))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	var urls []string
	if err := dec.Decode(&urls); err != nil {
		h.log.Error("DeleteURLs", logger.Error(err))
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		h.log.Error("DeleteURLs", logger.ErrorS("empty URL"))
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)

	h.svc.MakeDeleted(r.Context(), model.DeleteURLsRequest{
		UserID: userID,
		URLs:   urls,
	})
}

func (h *urlHandler) PingDB(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Ping(r.Context()); err != nil {
		h.log.Error("PingDB", logger.Error(err))
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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
