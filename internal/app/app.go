package app

import (
	"net/http"

	"shortener/internal/config"
	handler "shortener/internal/handler/http"
	"shortener/internal/repo"
	"shortener/internal/service"
	"shortener/internal/shared/compress/gzip"
	"shortener/internal/shared/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type urlHandler interface {
	ShortenURLText(w http.ResponseWriter, r *http.Request)
	ShortenURLJSON(w http.ResponseWriter, r *http.Request)
	RedirectURL(w http.ResponseWriter, r *http.Request)
}

type App struct {
	log *logger.Logger
	srv *http.Server
}

func New() *App {
	a := &App{}
	a.initDeps()

	return a
}

func (a *App) Run() error {
	defer a.log.Sync()

	a.log.Info(
		"Server started",
		logger.String("addr", a.srv.Addr),
	)
	err := a.srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		a.log.Error("server error", logger.Error(err))
		return err
	}
	return nil
}

func (a *App) initDeps() {
	cfg := config.MustLoad()
	log := logger.NewFile(cfg.App.LogLevel)
	repo := repo.NewURLRepository()
	svc := service.NewURLService(cfg.App.BaseAddr, repo)
	h := handler.NewURLHandler(log, svc)
	r := router(h)

	a.log = log
	a.srv = &http.Server{
		Addr:    cfg.App.Addr(),
		Handler: r,
	}

}

func router(h urlHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(logger.MiddlewareHTTP)
	r.Use(middleware.Recoverer)

	r.With(middleware.AllowContentType("text/plain", "text/html", "application/x-gzip"), gzip.Middleware).
		Post("/", h.ShortenURLText)
	r.With(middleware.AllowContentType("application/json"), gzip.Middleware).
		Post("/api/shorten", h.ShortenURLJSON)

	r.Get("/{short}", h.RedirectURL)

	return r
}
