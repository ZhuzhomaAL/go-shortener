package main

import (
	"github.com/ZhuzhomaAL/go-shortener/cmd/config"
	"github.com/ZhuzhomaAL/go-shortener/internal/app"
	"github.com/ZhuzhomaAL/go-shortener/internal/logger"
	"go.uber.org/zap"
	"log"
	"net/http"
)

func main() {
	appConfig := config.ParseFlags()
	if err := run(appConfig); err != nil {
		log.Fatal(err)
	}
}

func run(appConfig config.AppConfig) error {
	l := logger.MyLogger{}
	if err := l.Initialize(appConfig.FlagLogLevel); err != nil {
		return err
	}
	l.L.Info("Running server", zap.String("address", appConfig.FlagRunAddr))
	return http.ListenAndServe(appConfig.FlagRunAddr, app.Router(appConfig, l))
}
