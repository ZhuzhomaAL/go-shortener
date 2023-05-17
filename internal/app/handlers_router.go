package app

import (
	"github.com/ZhuzhomaAL/go-shortener/cmd/config"
	"github.com/go-chi/chi/v5"
	"sync"
)

var urlList sync.Map

func Router(appConfig config.AppConfig) chi.Router {
	urlList = sync.Map{}
	app := &app{
		appConfig: appConfig,
	}

	r := chi.NewRouter()
	r.Post("/", app.postHandler)
	r.Get("/{id}", app.getHandler)

	return r
}
