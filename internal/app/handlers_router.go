package app

import (
	"github.com/ZhuzhomaAL/go-shortener/cmd/config"
	"github.com/ZhuzhomaAL/go-shortener/internal/logger"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"sync"
)

var urlList sync.Map

func Router(appConfig config.AppConfig, logger logger.MyLogger) (chi.Router, error) {
	urlList = sync.Map{}
	fWriter, err := NewFileWriter(appConfig.FlagStorage)
	if err != nil {
		return nil, err
	}
	fReader, err := NewFileReader(appConfig.FlagStorage)
	if err != nil {
		return nil, err
	}
	app := &app{
		appConfig: appConfig,
		log:       logger,
		fWriter:   fWriter,
		fReader:   fReader,
	}
	for {
		url, err := app.fReader.ReadFile()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		urlList.Store(url.ShortURL, url.OriginalURL)
	}
	r := chi.NewRouter()
	r.Use(gzipMiddleware)
	r.Use(logger.RequestLogger)
	r.Post("/", app.postHandler)
	r.Post("/api/shorten", app.JSONHandler)
	r.Get(
		"/{id}", func(rw http.ResponseWriter, req *http.Request) {
			id := chi.URLParam(req, "id")
			app.getHandler(rw, req, id)
		},
	)

	return r, nil
}
