package app

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPostHandler_PositiveCases(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
		contentType    string
		URL            string
		baseURL        string
	}{
		{
			name:           "success_create_short_url",
			expectedStatus: http.StatusCreated,
			contentType:    "text/plain",
			URL:            "https://practicum.yandex.ru",
			baseURL:        "http://localhost:8080/",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				b := strings.NewReader(tt.URL)
				req := httptest.NewRequest(http.MethodPost, tt.baseURL, b)
				nr := httptest.NewRecorder()
				urlList = map[string]string{}
				postHandler(nr, req)
				res := nr.Result()
				assert.Equal(t, tt.expectedStatus, res.StatusCode, "Код ответа не совпадает с ожидаемым")
				assert.Contains(t, res.Header.Get("Content-Type"), tt.contentType, "Content-Type не совпадает с ожидаемым")
				result, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				err = res.Body.Close()
				require.NoError(t, err)
				require.NotEmpty(t, result, "Тело ответа пустое")
				require.True(t, strings.Contains(string(result), tt.baseURL))
				require.True(t, len(string(result)) > len(tt.baseURL))
			},
		)
	}
}

func TestPostHandler_NegativeCases(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
		baseURL        string
		expectedError  string
	}{
		{
			name:           "empty_body",
			expectedStatus: http.StatusBadRequest,
			baseURL:        "http://localhost:8080/",
			expectedError:  "response body is empty, expected not empty",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodPost, tt.baseURL, nil)
				nr := httptest.NewRecorder()
				urlList = map[string]string{}
				postHandler(nr, req)
				res := nr.Result()
				assert.Equal(t, tt.expectedStatus, res.StatusCode, "Код ответа не совпадает с ожидаемым")
				result, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				err = res.Body.Close()
				require.NoError(t, err)
				require.NotEmpty(t, result, "Тело ответа пустое")
				require.Contains(t, string(result), tt.expectedError, "Текст ошибки не совпадает с ожидаемым")
			},
		)
	}
}

func TestGetHandler_PositiveCases(t *testing.T) {
	tests := []struct {
		name             string
		expectedStatus   int
		endpoint         string
		shortURL         string
		expectedLocation string
	}{
		{
			name:             "success_short_url_id",
			expectedStatus:   http.StatusTemporaryRedirect,
			endpoint:         "http://localhost:8080/",
			shortURL:         "u9pEX2P5",
			expectedLocation: "https://practicum.yandex.ru",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodGet, tt.endpoint+tt.shortURL, nil)
				nr := httptest.NewRecorder()
				urlList = map[string]string{}
				urlList[tt.shortURL] = tt.expectedLocation
				getHandler(nr, req)
				res := nr.Result()
				assert.Equal(t, tt.expectedStatus, res.StatusCode, "Код ответа не совпадает с ожидаемым")
				result, err := ioutil.ReadAll(res.Body)
				require.NoError(t, err)
				err = res.Body.Close()
				require.NoError(t, err)
				require.NotEmpty(t, result, "Тело ответа пустое")
				require.Equal(t, string(result), tt.expectedLocation, "Location не совпадает с ожидаемым")
			},
		)
	}
}

func TestGetHandler_NegativeCases(t *testing.T) {
	tests := []struct {
		name             string
		expectedStatus   int
		endpoint         string
		shortURL         string
		expectedLocation string
		expectedError    string
	}{
		{
			name:           "empty_short_url_id",
			expectedStatus: http.StatusBadRequest,
			endpoint:       "http://localhost:8080/",
			expectedError:  "id is empty, expected not empty",
		},
		{
			name:           "location_not_found",
			expectedStatus: http.StatusBadRequest,
			endpoint:       "http://localhost:8080/",
			shortURL:       "LFGwsFFf",
			expectedError:  "location not found",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodGet, tt.endpoint+tt.shortURL, nil)
				nr := httptest.NewRecorder()
				urlList = map[string]string{}
				getHandler(nr, req)
				res := nr.Result()
				assert.Equal(t, tt.expectedStatus, res.StatusCode, "Код ответа не совпадает с ожидаемым")
				result, err := ioutil.ReadAll(res.Body)
				require.NoError(t, err)
				err = res.Body.Close()
				require.NoError(t, err)
				require.NotEmpty(t, result, "Тело ответа пустое")
				require.Contains(t, string(result), tt.expectedError, "Текст ошибки не совпадает с ожидаемым")
			},
		)
	}
}
