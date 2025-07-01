package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"slices"
)

type gzipReader struct {
	r  io.ReadCloser
	gz *gzip.Reader
}

func newGzipReader(r io.ReadCloser) (*gzipReader, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &gzipReader{r: r, gz: gz}, nil
}

func (gr *gzipReader) Read(p []byte) (int, error) {
	return gr.gz.Read(p)
}

func (gr *gzipReader) Close() error {
	if err := gr.r.Close(); err != nil {
		return err
	}

	return gr.gz.Close()
}

type gzipWriter struct {
	w  http.ResponseWriter
	gz *gzip.Writer
}

func newGzipWriter(w http.ResponseWriter) *gzipWriter {
	return &gzipWriter{w: w, gz: gzip.NewWriter(w)}
}

func (gw *gzipWriter) Header() http.Header {
	return gw.w.Header()
}

func (gw *gzipWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		gw.w.Header().Set("Content-Encoding", "gzip")
	}
	gw.w.WriteHeader(statusCode)
}

func (gw *gzipWriter) Write(p []byte) (int, error) {
	return gw.gz.Write(p)
}

func (gw *gzipWriter) Close() error {
	return gw.gz.Close()
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		if slices.Contains(r.Header.Values("Content-Encoding"), "gzip") {
			gr, err := newGzipReader(r.Body)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			defer gr.Close()

			r.Body = gr
		}
		if slices.Contains(r.Header.Values("Accept-Encoding"), "gzip") {
			gw := newGzipWriter(w)
			defer gw.Close()
			ow = gw
		}

		next.ServeHTTP(ow, r)
	})
}
