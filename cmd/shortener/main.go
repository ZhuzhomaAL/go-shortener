package main

import (
	"database/sql"
	"github.com/ZhuzhomaAL/go-shortener/cmd/config"
	"github.com/ZhuzhomaAL/go-shortener/internal/app"
	"github.com/ZhuzhomaAL/go-shortener/internal/logger"
	"go.uber.org/zap"
	"log"
	"net/http"
)

func main() {
	appConfig := config.ParseFlags()

	db := app.GetConnection(appConfig.FlagDB)

	if db != nil {
		defer db.Close()
	}

	if err := run(appConfig, db); err != nil {
		log.Fatal(err)
	}
}

func run(appConfig config.AppConfig, db *sql.DB) error {
	l, err := logger.Initialize(appConfig.FlagLogLevel)
	if err != nil {
		return err
	}
	l.L.Info("Running server", zap.String("address", appConfig.FlagRunAddr))
	r, err := app.Router(appConfig, l, db)
	if err != nil {
		return err
	}
	if db != nil {
		err = app.CreateShortURLTable(db)
		if err != nil {
			l.L.Error("failed to create short url table", zap.Error(err))
		}
		err = app.CreateIndex(db)
		if err != nil {
			l.L.Error("failed to create short url index", zap.Error(err))
		}
	}

	return http.ListenAndServe(appConfig.FlagRunAddr, r)
}
