package app

import (
	"database/sql"
	"fmt"
	"github.com/ZhuzhomaAL/go-shortener/cmd/config"
	"github.com/ZhuzhomaAL/go-shortener/internal/logger"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"io"
	"net/http"
	"sync"
)

var urlList sync.Map

func Router(appConfig config.AppConfig, logger logger.MyLogger, db *sql.DB) (chi.Router, error) {

	urlList = sync.Map{}
	app := &app{
		appConfig: appConfig,
		log:       logger,
		fWriter:   nil,
		fReader:   nil,
		db:        db,
	}
	if appConfig.FlagStorage != "" {
		fWriter, err := NewFileWriter(appConfig.FlagStorage)
		if err != nil {
			return nil, err
		}
		fReader, err := NewFileReader(appConfig.FlagStorage)
		if err != nil {
			return nil, err
		}

		app.fWriter = fWriter
		app.fReader = fReader

		for {
			url, err := app.fReader.ReadFile()
			if err != nil {
				if err == io.EOF {
					break
				}
				return nil, fmt.Errorf("failed to read the storage file: %w", err)

			}
			urlList.Store(url.ShortURL, url.OriginalURL)
		}

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
	r.Get("/ping", app.pingDBHandler)

	return r, nil
}
