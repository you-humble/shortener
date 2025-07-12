package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"shortener/internal/model"
	"shortener/internal/shared/logger"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	addr string = "localhost:8080"
	good string = "good"
)

type urlServiceMock struct{}

// TODO: test json, test gzip, test ping
func (s *urlServiceMock) GenerateShortURL(ctx context.Context, scheme, userID, original string) (string, error) {
	if original == "wrong" {
		return "", errors.New("service error")
	}
	return fmt.Sprintf("http://%s/%s", addr, good), nil
}

func (s *urlServiceMock) OriginalURL(ctx context.Context, short string) (string, error) {
	if short == "" {
		return "", errors.New("service error")
	}
	return good, nil
}

func (s *urlServiceMock) Ping(ctx context.Context) error { return nil }

func (s *urlServiceMock) GenerateShortBatch(
	ctx context.Context,
	scheme string,
	userID string,
	elems []model.ShortenBatchRequest,
) ([]model.ShortenBatchResponse, error) {
	return []model.ShortenBatchResponse{}, nil
}

func (s *urlServiceMock) UserStore(context.Context, string) ([]model.URLStore, error) {
	return nil, nil
}

func (s *urlServiceMock) MakeDeleted(context.Context, model.DeleteURLsRequest) {}

func (s *urlServiceMock) URLByID(context.Context, int) (model.URLStore, error) {
	return model.URLStore{}, nil
}

type authServiceMock struct{}

func (s *authServiceMock) UserIDFromContext(context.Context) (string, bool) {
	return "1", true
}

func TestURLHandler(t *testing.T) {
	type want struct {
		statusCode int
		mediaType  string
		resp       string
	}

	testCases := []struct {
		name          string
		query         string
		body          string
		wantPOST      want
		getStatusCode int
	}{
		{
			name:  "shortenURL success",
			query: "/",
			body:  "https://www.google.com",
			wantPOST: want{
				statusCode: http.StatusCreated,
				mediaType:  "text/plain",
				resp:       fmt.Sprintf("http://%s/%s", addr, good),
			},
			getStatusCode: http.StatusTemporaryRedirect,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewURLHandler(logger.L(), &urlServiceMock{}, &authServiceMock{})

			w := httptest.NewRecorder()
			body := bytes.NewBuffer([]byte(tc.body))
			r := httptest.NewRequest(http.MethodPost, tc.query, body)
			r.Header.Add("Content-Type", "text/plain")

			h.ShortenURLText(w, r)

			res := w.Result()

			assert.Equal(t, tc.wantPOST.statusCode, res.StatusCode)
			assert.Contains(t, res.Header.Get("Content-Type"), tc.wantPOST.mediaType)

			b, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			require.NoError(t, res.Body.Close())

			shortURL := string(b)
			assert.Equal(t, tc.wantPOST.resp, shortURL)

			log.Println(shortURL)
			w = httptest.NewRecorder()
			r = httptest.NewRequest(http.MethodGet, shortURL, nil)

			h.RedirectURL(w, r)

			res = w.Result()

			assert.Equal(t, tc.getStatusCode, res.StatusCode)
		})
	}
}
