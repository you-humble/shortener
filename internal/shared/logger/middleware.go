package logger

import (
	"net/http"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	respData *struct {
		status int
		size   int
	}
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.respData.size += size
	return size, err
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.respData.status = statusCode
}

func MiddlewareHTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lw := loggingResponseWriter{
			ResponseWriter: w,
			respData: &struct {
				status int
				size   int
			}{0, 0},
		}

		next.ServeHTTP(&lw, r)

		L().Info("",
			String("method", r.Method),
			String("uri", r.RequestURI),
			Int("status", lw.respData.status),
			Duration("time", time.Since(start)),
			Int("size", lw.respData.size),
		)
	})
}
