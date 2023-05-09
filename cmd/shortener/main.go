package main

import (
	"github.com/ZhuzhomaAL/go-shortener/internal/app"
	"net/http"
)

func main() {
	http.ListenAndServe(":8080", app.Router())
}
