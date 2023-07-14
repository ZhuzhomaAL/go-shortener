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

	db, err := sql.Open("postgres", appConfig.FlagDB)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
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
	return http.ListenAndServe(appConfig.FlagRunAddr, r)
}
