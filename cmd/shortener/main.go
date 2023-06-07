package main

import (
	"github.com/ZhuzhomaAL/go-shortener/cmd/config"
	"github.com/ZhuzhomaAL/go-shortener/internal/app"
	"github.com/ZhuzhomaAL/go-shortener/internal/logger"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
)

func main() {
	appConfig := config.ParseFlags()
	if err := run(appConfig); err != nil {
		log.Fatal(err)
	}
}

func run(appConfig config.AppConfig) error {
	l, err := logger.Initialize(appConfig.FlagLogLevel)
	if err != nil {
		return err
	}
	err = os.MkdirAll("tmp", 0750)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}
	l.L.Info("Running server", zap.String("address", appConfig.FlagRunAddr))
	r, err := app.Router(appConfig, l)
	if err != nil {
		return err
	}
	return http.ListenAndServe(appConfig.FlagRunAddr, r)
}
