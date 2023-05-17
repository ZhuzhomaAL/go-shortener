package app

import (
	"fmt"
	"github.com/ZhuzhomaAL/go-shortener/cmd/config"
	"github.com/dchest/uniuri"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"net/url"
)

type app struct {
	appConfig config.AppConfig
}

func (a *app) postHandler(rw http.ResponseWriter, req *http.Request) {
	if req == nil {
		http.Error(rw, "request is empty, expected not empty", http.StatusBadRequest)
		return
	}
	rw.Header().Set("Content-Type", "text/plain")
	resp, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(rw, "failed to process request", http.StatusBadRequest)
		return
	}
	if len(resp) == 0 {
		http.Error(rw, "response body is empty, expected not empty", http.StatusBadRequest)
		return
	}
	genShortStr := uniuri.NewLen(8)
	urlList.Store(genShortStr, string(resp))
	respString, err := url.JoinPath(a.appConfig.FlagShortAddr, genShortStr)
	if err != nil {
		http.Error(rw, "failed to process request", http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusCreated)
	if _, err := rw.Write([]byte(respString)); err != nil {
		return
	}
}

func (a *app) getHandler(rw http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")
	location, ok := urlList.Load(id)
	locationStr := fmt.Sprintf("%v", location)
	if !ok {
		http.Error(rw, "location not found", http.StatusBadRequest)
		return
	}
	rw.Header().Set("Location", locationStr)
	rw.WriteHeader(http.StatusTemporaryRedirect)
	if _, err := rw.Write([]byte(locationStr)); err != nil {
		return
	}
}
