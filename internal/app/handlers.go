package app

import (
	"encoding/json"
	"fmt"
	"github.com/ZhuzhomaAL/go-shortener/cmd/config"
	"github.com/ZhuzhomaAL/go-shortener/internal/logger"
	"github.com/dchest/uniuri"
	"io"
	"log"
	"net/http"
	"net/url"
)

type app struct {
	appConfig config.AppConfig
	log       logger.MyLogger
}

type result struct {
	Result string `json:"result"`
}

type reqURL struct {
	ReqURL string `json:"url"`
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
		log.Println(err)
		return
	}
}

func (a *app) getHandler(rw http.ResponseWriter, req *http.Request, id string) {
	location, ok := urlList.Load(id)
	locationStr := fmt.Sprintf("%v", location)
	if !ok {
		http.Error(rw, "location not found", http.StatusBadRequest)
		return
	}
	rw.Header().Set("Location", locationStr)
	rw.WriteHeader(http.StatusTemporaryRedirect)
	if _, err := rw.Write([]byte(locationStr)); err != nil {
		log.Println(err)
		return
	}
}

func (a *app) JSONHandler(rw http.ResponseWriter, req *http.Request) {
	var reqUrl reqURL
	var result result

	if req.Body == nil {
		http.Error(rw, "request is empty, expected not empty", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(req.Body).Decode(&reqUrl); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	genShortStr := uniuri.NewLen(8)
	urlList.Store(genShortStr, reqUrl.ReqURL)
	respString, err := url.JoinPath(a.appConfig.FlagShortAddr, genShortStr)
	if err != nil {
		http.Error(rw, "failed to process request", http.StatusBadRequest)
		return
	}
	result.Result = respString
	resp, err := json.Marshal(result)
	if err != nil {
		http.Error(rw, "failed to process request", http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	if _, err := rw.Write(resp); err != nil {
		log.Println(err)
		return
	}
}
