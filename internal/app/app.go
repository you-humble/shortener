package app

import (
	"context"
	"net/http"

	"shortener/internal/config"
	handler "shortener/internal/handler/http"
	frepo "shortener/internal/repo/file"
	mrepo "shortener/internal/repo/memory"
	"shortener/internal/repo/pg"
	"shortener/internal/service"
	"shortener/internal/shared/compress/gzip"
	"shortener/internal/shared/database/postgres"
	"shortener/internal/shared/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type urlHandler interface {
	ShortenURLText(w http.ResponseWriter, r *http.Request)
	ShortenURLJSON(w http.ResponseWriter, r *http.Request)
	ShortenBatchJSON(w http.ResponseWriter, r *http.Request)
	RedirectURL(w http.ResponseWriter, r *http.Request)
	AllUserURLs(w http.ResponseWriter, r *http.Request)
	DeleteURLs(w http.ResponseWriter, r *http.Request)
	PingDB(w http.ResponseWriter, r *http.Request)
	URLByID(w http.ResponseWriter, r *http.Request)
}

type Registrator interface {
	CheckInMiddleware(http.Handler) http.Handler
}

type App struct {
	log    *logger.Logger
	srv    *http.Server
	cancel context.CancelFunc
}

func New() *App {
	a := &App{}
	a.initDeps()

	return a
}

func (a *App) Run() error {
	defer a.log.Sync()
	defer a.cancel()

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
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel

	cfg := config.MustLoad()
	log := logger.New(cfg.App.LogLevel)

	var repo service.URLRepository
	var err error
	if cfg.DB.DSN != "" {
		db, err := postgres.NewConnect(ctx, cfg.DB.DSN)
		if err != nil {
			logger.Fatal("", logger.ErrorS(err.Error()))
		}
		log.Info("The database is connected")

		repo, err = pg.NewURLRepository(ctx, db)
		if err != nil {
			logger.Fatal("new postgres storage", logger.ErrorS(err.Error()))
		}
		log.Info("Using postgres storage")
	} else if cfg.DB.FileStorage != "" {
		repo, err = frepo.NewURLRepository(cfg.DB.FileStorage)
		if err != nil {
			logger.Fatal("failed to create new file repository")
		}
		log.Info("Using file storage")
	} else {
		repo, err = mrepo.NewURLRepository()
		if err != nil {
			logger.Fatal("failed to create new in-memory repository")
		}
		log.Info("Using in-memory storage")
	}

	urlSvc := service.NewURLService(ctx, cfg.App.BaseAddr, repo)
	authSvc := service.NewAuthService(log, cfg.Auth.Secret, cfg.Auth.TokenExpire)
	h := handler.NewURLHandler(log, urlSvc, authSvc)
	r := router(h, authSvc)

	a.log = log
	a.srv = &http.Server{
		Addr:    cfg.App.Addr(),
		Handler: r,
	}
}

func router(h urlHandler, reg Registrator) http.Handler {
	r := chi.NewRouter()

	r.Use(logger.MiddlewareHTTP)
	r.Use(middleware.Recoverer)
	r.Use(reg.CheckInMiddleware)

	r.With(middleware.AllowContentType("text/plain", "text/html", "application/x-gzip"), gzip.Middleware).
		Post("/", h.ShortenURLText)
	r.With(middleware.AllowContentType("application/json"), gzip.Middleware).
		Post("/api/shorten", h.ShortenURLJSON)
	r.With(middleware.AllowContentType("application/json"), gzip.Middleware).
		Post("/api/shorten/batch", h.ShortenBatchJSON)

	r.Get("/{short}", h.RedirectURL)
	r.Get("/{id:[0-9]+}", h.URLByID)
	r.Get("/api/user/urls", h.AllUserURLs)
	r.Get("/ping", h.PingDB)

	r.Delete("/api/user/urls", h.DeleteURLs)

	return r
}
