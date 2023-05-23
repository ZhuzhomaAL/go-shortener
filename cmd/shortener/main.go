package main

import (
	"github.com/ZhuzhomaAL/go-shortener/cmd/config"
	"github.com/ZhuzhomaAL/go-shortener/internal/app"
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
	log.Println("Running server on", appConfig.FlagRunAddr)
	return http.ListenAndServe(appConfig.FlagRunAddr, app.Router(appConfig))
}
