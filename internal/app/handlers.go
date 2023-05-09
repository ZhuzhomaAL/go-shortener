package app

import (
	"fmt"
	"github.com/dchest/uniuri"
	"io"
	"net/http"
	"strings"
)

func handler(rw http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		getHandler(rw, req)
	case http.MethodPost:
		postHandler(rw, req)
	}
}

func postHandler(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if req == nil {
		http.Error(rw, "request is empty, expected not empty", http.StatusBadRequest)
		return
	}
	rw.Header().Set("Content-Type", "text/plain")
	resp, err := io.ReadAll(req.Body)
	if len(resp) == 0 {
		http.Error(rw, "response body is empty, expected not empty", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(rw, "failed to process request", http.StatusBadRequest)
		return
	}
	genShortStr := uniuri.NewLen(8)
	urlList[genShortStr] = string(resp)
	respString := fmt.Sprintf("http://localhost:8080/%s", genShortStr)
	rw.WriteHeader(http.StatusCreated)
	_, err = rw.Write([]byte(respString))
	if err != nil {
		return
	}
}

func getHandler(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := req.URL.Path
	id = strings.TrimPrefix(id, "/")
	if id == "" {
		http.Error(rw, "id is empty, expected not empty", http.StatusBadRequest)
		return
	}
	location, ok := urlList[id]
	if !ok {
		http.Error(rw, "location not found", http.StatusBadRequest)
		return
	}
	rw.Header().Set("Location", location)
	rw.WriteHeader(http.StatusTemporaryRedirect)
	_, err := rw.Write([]byte(location))
	if err != nil {
		return
	}
}
