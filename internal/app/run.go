package app

import (
	"net/http"
)

var urlList map[string]string

func Run() error {
	urlList = map[string]string{}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	return http.ListenAndServe(`:8080`, mux)
}
