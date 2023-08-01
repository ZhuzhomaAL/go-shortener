package app

import (
	"encoding/json"
	"github.com/ZhuzhomaAL/go-shortener/internal/store"
	"github.com/ZhuzhomaAL/go-shortener/internal/utils"
	"github.com/dchest/uniuri"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"net/url"
)

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

	request, err := io.ReadAll(req.Body)
	if err != nil {
		a.myLogger.L.Error("failed to process request", zap.Error(err))
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		return
	}
	if len(request) == 0 {
		http.Error(rw, "response body is empty, expected not empty", http.StatusBadRequest)
		return
	}
	genShortStr := uniuri.NewLen(8)
	userID, ok := req.Context().Value(utils.ContextUserID).(uuid.UUID)
	if !ok {
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		return
	}
	URL := store.URL{
		OriginalURL: string(request),
		ShortURL:    genShortStr,
		UserID:      userID,
	}
	err = a.writer.SaveURL(req.Context(), URL)
	if err != nil {
		if err, ok := err.(*store.ConflictError); ok {
			a.myLogger.L.Error("duplicate key value", zap.Error(err))
			a.makeSinglePlainResponse(rw, err.ShortURL, http.StatusConflict)
			return
		}
		a.myLogger.L.Error("failed to persist data", zap.Error(err))
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		return
	}
	a.makeSinglePlainResponse(rw, genShortStr, http.StatusCreated)
}

func (a *app) makeSinglePlainResponse(rw http.ResponseWriter, genShortStr string, status int) {
	respString, err := url.JoinPath(a.appConfig.FlagShortAddr, genShortStr)
	if err != nil {
		a.myLogger.L.Error("failed to process request", zap.Error(err))
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(status)
	if _, err := rw.Write([]byte(respString)); err != nil {
		log.Println(err)
		return
	}
}

func (a *app) getHandler(rw http.ResponseWriter, req *http.Request, id string) {
	location, err := a.reader.GetURL(req.Context(), id)
	if err != nil {
		if err, ok := err.(*store.DeletedURLError); ok {
			a.myLogger.L.Error("requested URL deleted", zap.Error(err))
			rw.WriteHeader(http.StatusGone)
			return
		}
		http.Error(rw, "location not found", http.StatusBadRequest)
		return
	}
	rw.Header().Set("Location", location)
	rw.WriteHeader(http.StatusTemporaryRedirect)
	if _, err := rw.Write([]byte(location)); err != nil {
		log.Println(err)
		return
	}
}

func (a *app) pingDBHandler(rw http.ResponseWriter, req *http.Request) {
	if reader, ok := a.reader.(store.PingableReader); ok {
		err := reader.Ping(req.Context())
		if err != nil {
			a.myLogger.L.Error("failed to connect to database", zap.Error(err))
			http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		}
		rw.WriteHeader(http.StatusOK)
		return
	}
	a.myLogger.L.Error("failed to connect to database")
	http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
}

type batchRes struct {
	ID       string `json:"correlation_id"`
	ShortURL string `json:"short_url"`
}

type batchURL struct {
	ID          string `json:"correlation_id"`
	OriginalURL string `json:"original_url"`
	ShortURL    string
}

func (a *app) batchHandler(rw http.ResponseWriter, req *http.Request) {

	dec := json.NewDecoder(req.Body)

	var batchURL []batchURL
	for dec.More() {
		err := dec.Decode(&batchURL)
		if err != nil {
			if err == io.EOF {
				http.Error(rw, "request is empty, expected not empty", http.StatusBadRequest)
				return
			}
			a.myLogger.L.Error("failed to decode request", zap.Error(err))
			http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
			return
		}

	}
	userID, ok := req.Context().Value(utils.ContextUserID).(uuid.UUID)
	if !ok {
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		return
	}
	var URLs []store.URL
	for i := range batchURL {
		batchURL[i].ShortURL = uniuri.NewLen(8)
		URL := store.URL{
			OriginalURL: batchURL[i].OriginalURL,
			ShortURL:    batchURL[i].ShortURL,
			UserID:      userID,
		}
		URLs = append(URLs, URL)

	}
	err := a.writer.SaveBatch(req.Context(), URLs)
	if err != nil {
		a.myLogger.L.Error("failed to persist data", zap.Error(err))
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		return
	}

	var result []batchRes
	for _, URL := range batchURL {
		var resItem batchRes
		resItem.ID = URL.ID
		respString, err := url.JoinPath(a.appConfig.FlagShortAddr, URL.ShortURL)
		if err != nil {
			a.myLogger.L.Error("failed to process request", zap.Error(err))
			http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
			return
		}
		resItem.ShortURL = respString
		result = append(result, resItem)
	}
	resp, err := json.Marshal(result)
	if err != nil {
		a.myLogger.L.Error("failed to process request", zap.Error(err))
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	if _, err := rw.Write(resp); err != nil {
		a.myLogger.L.Error("failed to retrieve response", zap.Error(err))
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		return
	}
}

func (a *app) JSONHandler(rw http.ResponseWriter, req *http.Request) {
	var reqURL reqURL

	if err := json.NewDecoder(req.Body).Decode(&reqURL); err != nil {
		if err == io.EOF {
			http.Error(rw, "request is empty, expected not empty", http.StatusBadRequest)
			return
		}
		a.myLogger.L.Error("failed to decode request", zap.Error(err))
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		return
	}

	genShortStr := uniuri.NewLen(8)
	userID, ok := req.Context().Value(utils.ContextUserID).(uuid.UUID)
	if !ok {
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		return
	}
	URL := store.URL{
		OriginalURL: reqURL.ReqURL,
		ShortURL:    genShortStr,
		UserID:      userID,
	}
	err := a.writer.SaveURL(req.Context(), URL)
	if err != nil {
		if err, ok := err.(*store.ConflictError); ok {
			a.myLogger.L.Error("duplicate key value", zap.Error(err))
			a.makeSingleJSONResponse(rw, err.ShortURL, http.StatusConflict)
			return
		}
		a.myLogger.L.Error("failed to persist data", zap.Error(err))
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		return
	}

	a.makeSingleJSONResponse(rw, genShortStr, http.StatusCreated)
}

func (a *app) makeSingleJSONResponse(rw http.ResponseWriter, genShortStr string, status int) {
	respString, err := url.JoinPath(a.appConfig.FlagShortAddr, genShortStr)
	if err != nil {
		a.myLogger.L.Error("failed to process request", zap.Error(err))
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		return
	}
	var result result

	result.Result = respString
	resp, err := json.Marshal(result)
	if err != nil {
		a.myLogger.L.Error("failed to process request", zap.Error(err))
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	if _, err := rw.Write(resp); err != nil {
		a.myLogger.L.Error("failed to retrieve response", zap.Error(err))
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		return
	}
}

type usersURL struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
}

func (a *app) getUserURLHandler(rw http.ResponseWriter, req *http.Request) {
	reader, ok := a.reader.(store.UserIDReader)
	if !ok {
		a.myLogger.L.Error("reader can not read user ID")
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
	}
	userID, ok := req.Context().Value(utils.ContextUserID).(uuid.UUID)
	if !ok {
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		return
	}
	urls, err := reader.GetURLsByUserID(req.Context(), userID.String())
	if err != nil {
		a.myLogger.L.Error("failed to get URLs by user ID", zap.Error(err))
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		return
	}
	if len(urls) == 0 {
		rw.WriteHeader(http.StatusNoContent)
		return
	}
	var usersURLs []usersURL
	for _, URL := range urls {
		shortURL, err := url.JoinPath(a.appConfig.FlagShortAddr, URL.ShortURL)
		if err != nil {
			a.myLogger.L.Error("failed to process request", zap.Error(err))
			http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
			return
		}
		userURL := usersURL{
			OriginalURL: URL.OriginalURL,
			ShortURL:    shortURL,
		}
		usersURLs = append(usersURLs, userURL)
	}
	resp, err := json.Marshal(usersURLs)
	if err != nil {
		a.myLogger.L.Error("failed to process request", zap.Error(err))
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	if _, err := rw.Write(resp); err != nil {
		a.myLogger.L.Error("failed to retrieve response", zap.Error(err))
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		return
	}
}

func (a *app) deleteHandler(rw http.ResponseWriter, req *http.Request) {
	var result []string
	if err := json.NewDecoder(req.Body).Decode(&result); err != nil {
		if err == io.EOF {
			http.Error(rw, "request is empty, expected not empty", http.StatusBadRequest)
			return
		}
		a.myLogger.L.Error("failed to decode request", zap.Error(err))
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		return
	}
	userID, ok := req.Context().Value(utils.ContextUserID).(uuid.UUID)
	if !ok {
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
		return
	}
	reader, ok := a.reader.(store.UserIDReader)
	if !ok {
		a.myLogger.L.Error("reader can not read user ID")
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
	}

	var shortUrls []store.URL
	for _, res := range result {
		URL := store.URL{
			ShortURL: res,
		}
		shortUrls = append(shortUrls, URL)
	}
	filteredURLs, err := reader.FilterURLsByUserID(req.Context(), userID.String(), shortUrls)
	if err != nil {
		a.myLogger.L.Error("can not filter urls by user ID")
		http.Error(rw, "internal server error occurred", http.StatusInternalServerError)
	}
	for _, fu := range filteredURLs {
		a.storeChan <- fu
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusAccepted)
}
