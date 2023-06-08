package app

import (
	"encoding/json"
	"fmt"
	"github.com/ZhuzhomaAL/go-shortener/cmd/config"
	"github.com/ZhuzhomaAL/go-shortener/internal/logger"
	"github.com/dchest/uniuri"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"net/url"
)

type app struct {
	appConfig config.AppConfig
	log       logger.MyLogger
	fWriter   *Writer
	fReader   *Reader
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
		a.log.L.Error("failed to process request")
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		return
	}
	if len(resp) == 0 {
		http.Error(rw, "response body is empty, expected not empty", http.StatusBadRequest)
		return
	}
	genShortStr := uniuri.NewLen(8)
	urlList.Store(genShortStr, string(resp))
	id := uuid.New()
	fileURL := &URL{
		id,
		genShortStr,
		string(resp),
	}
	err = a.fWriter.WriteFile(fileURL)
	if err != nil {
		a.log.L.Error("failed to persist data")
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		return
	}
	respString, err := url.JoinPath(a.appConfig.FlagShortAddr, genShortStr)
	if err != nil {
		a.log.L.Error("failed to process request")
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
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
	var reqURL reqURL
	var result result

	if err := json.NewDecoder(req.Body).Decode(&reqURL); err != nil {
		if err == io.EOF {
			http.Error(rw, "request is empty, expected not empty", http.StatusBadRequest)
			return
		}
	}

	genShortStr := uniuri.NewLen(8)
	urlList.Store(genShortStr, reqURL.ReqURL)
	respString, err := url.JoinPath(a.appConfig.FlagShortAddr, genShortStr)
	if err != nil {
		a.log.L.Error("failed to process request")
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		return
	}
	id := uuid.New()
	fileURL := &URL{
		id,
		genShortStr,
		reqURL.ReqURL,
	}
	err = a.fWriter.WriteFile(fileURL)
	if err != nil {
		a.log.L.Error("failed to persist data")
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		return
	}
	result.Result = respString
	resp, err := json.Marshal(result)
	if err != nil {
		a.log.L.Error("failed to process request")
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	if _, err := rw.Write(resp); err != nil {
		a.log.L.Error("failed to retrieve response")
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		return
	}
}
