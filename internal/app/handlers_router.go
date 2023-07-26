package app

import (
	"github.com/ZhuzhomaAL/go-shortener/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	"net/http"
	"time"
)

func Router(app *app) (chi.Router, error) {
	r := chi.NewRouter()
	r.Use(utils.GzipMiddleware)
	r.Use(app.myLogger.RequestLogger)
	r.Use(utils.AuthMiddleware)
	r.Use(middleware.Timeout(10 * time.Second))
	r.Post("/", app.postHandler)
	r.Route(
		"/api/shorten", func(r chi.Router) {
			r.Post("/", app.JSONHandler)
			r.Post("/batch", app.batchHandler)
		},
	)
	r.Get(
		"/{id}", func(rw http.ResponseWriter, req *http.Request) {
			id := chi.URLParam(req, "id")
			app.getHandler(rw, req, id)
		},
	)
	r.Get("/ping", app.pingDBHandler)
	r.Get("/api/user/urls", app.getUserURLHandler)

	return r, nil
}
