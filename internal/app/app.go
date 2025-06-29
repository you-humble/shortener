package app

import (
	"log"
	"net/http"

	"shortener/internal/config"
	handler "shortener/internal/handler/http"
	"shortener/internal/repo"
	"shortener/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type urlHandler interface {
	ShortenURLText(w http.ResponseWriter, r *http.Request)
	ShortenURLJSON(w http.ResponseWriter, r *http.Request)
	RedirectURL(w http.ResponseWriter, r *http.Request)
}

type App struct {
	srv *http.Server
}

func New() *App {
	a := &App{}
	a.initDeps()

	return a
}

func (a *App) Run() error {
	log.Printf("server started on address - %s", a.srv.Addr)
	err := a.srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (a *App) initDeps() {
	cfg := config.MustLoad()
	repo := repo.NewURLRepository()
	svc := service.NewURLService(cfg.App.BaseAddr, repo)
	h := handler.NewURLHandler(svc)
	r := router(h)

	a.srv = &http.Server{
		Addr:    cfg.App.Addr(),
		Handler: r,
	}
}

func router(h urlHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Group(func(r chi.Router) {
		r.Use(middleware.AllowContentType("text/plain", "application/json"))
		r.Post("/", h.ShortenURLText)
		r.Post("/api/shorten", h.ShortenURLJSON)
	})
	r.Get("/{short}", h.RedirectURL)

	return r
}
