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

	db := getConnection(appConfig.FlagDB)

	if db != nil {
		defer db.Close()
	}

	if err := run(appConfig, db); err != nil {
		log.Fatal(err)
	}
}

func getConnection(dsnString string) *sql.DB {
	if dsnString == "" {
		return nil
	}

	db, err := sql.Open("postgres", dsnString)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	return db
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
