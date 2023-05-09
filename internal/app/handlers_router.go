package app

import (
	"github.com/go-chi/chi/v5"
)

var urlList map[string]string

func Router() chi.Router {
	urlList = map[string]string{}

	r := chi.NewRouter()
	r.Post("/", postHandler)
	r.Get("/{id}", getHandler)

	return r
}
