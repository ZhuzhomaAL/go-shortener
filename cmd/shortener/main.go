package main

import (
	"github.com/ZhuzhomaAL/go-shortener/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		panic(err)
	}
}
