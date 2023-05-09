package main

import (
	"fmt"
	"github.com/ZhuzhomaAL/go-shortener/cmd/config"
	"github.com/ZhuzhomaAL/go-shortener/internal/app"
	"net/http"
)

func main() {
	config.ParseFlags()
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	fmt.Println("Running server on", config.FlagRunAddr)
	return http.ListenAndServe(config.FlagRunAddr, app.Router())
}
