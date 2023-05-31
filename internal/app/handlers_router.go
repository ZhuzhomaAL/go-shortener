package app

import (
	"github.com/ZhuzhomaAL/go-shortener/cmd/config"
	"github.com/ZhuzhomaAL/go-shortener/internal/logger"
	"github.com/go-chi/chi/v5"
	"net/http"
	"sync"
)

var urlList sync.Map

func Router(appConfig config.AppConfig, log logger.MyLogger) chi.Router {
	urlList = sync.Map{}
	app := &app{
		appConfig: appConfig,
		log:       log,
	}

	r := chi.NewRouter()
	r.Use(log.RequestLogger)
	r.Post("/", app.postHandler)
	r.Post("/api/shorten", app.JSONHandler)
	r.Get(
		"/{id}", func(rw http.ResponseWriter, req *http.Request) {
			id := chi.URLParam(req, "id")
			app.getHandler(rw, req, id)
		},
	)

	return r
}
