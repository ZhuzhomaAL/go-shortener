package app

import (
	"github.com/ZhuzhomaAL/go-shortener/cmd/config"
	"github.com/ZhuzhomaAL/go-shortener/internal/logger"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
)

var ts *httptest.Server

func TestMain(m *testing.M) {
	appConfig := config.ParseFlags()
	l := logger.MyLogger{}
	err := l.Initialize(appConfig.FlagLogLevel)
	if err != nil {
		return
	}
	ts = httptest.NewServer(Router(appConfig, l))
	defer ts.Close()
	status := m.Run()
	os.Exit(status)
}
func testRequest(t *testing.T, ts *httptest.Server, method, path string, body string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, strings.NewReader(body))
	require.NoError(t, err)
	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

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
		tt := tt
		t.Run(
			tt.name, func(t *testing.T) {
				urlList = sync.Map{}
				resp, respBody := testRequest(t, ts, "POST", "/", tt.URL)
				defer resp.Body.Close()
				assert.Equal(t, tt.expectedStatus, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
				assert.Contains(t, resp.Header.Get("Content-Type"), tt.contentType, "Content-Type не совпадает с ожидаемым")
				require.NotEmpty(t, respBody, "Тело ответа пустое")
				require.True(t, strings.Contains(respBody, tt.baseURL))
				require.True(t, len(respBody) > len(tt.baseURL))
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
		tt := tt
		t.Run(
			tt.name, func(t *testing.T) {
				urlList = sync.Map{}
				resp, respBody := testRequest(t, ts, "POST", "/", "")
				defer resp.Body.Close()
				assert.Equal(t, tt.expectedStatus, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
				require.NotEmpty(t, respBody, "Тело ответа пустое")
				require.Contains(t, respBody, tt.expectedError, "Текст ошибки не совпадает с ожидаемым")
			},
		)
	}
}

func TestGetHandler_PositiveCases(t *testing.T) {
	tests := []struct {
		name             string
		expectedStatus   int
		shortURL         string
		expectedLocation string
	}{
		{
			name:             "success_short_url_id",
			expectedStatus:   http.StatusOK,
			shortURL:         "u9pEX2P5",
			expectedLocation: "https://www.google.com/",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(
			tt.name, func(t *testing.T) {
				urlList = sync.Map{}
				urlList.Store(tt.shortURL, tt.expectedLocation)
				resp, respBody := testRequest(t, ts, "GET", "/"+tt.shortURL, "")
				defer resp.Body.Close()
				assert.Equal(t, tt.expectedStatus, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
				require.NotEmpty(t, respBody, "Тело ответа пустое")
				require.Contains(t, respBody, tt.expectedLocation, "Location не совпадает с ожидаемым")
			},
		)
	}
}

func TestGetHandler_NegativeCases(t *testing.T) {
	tests := []struct {
		name             string
		expectedStatus   int
		shortURL         string
		expectedLocation string
		expectedError    string
		wantError        bool
	}{
		{
			name:           "empty_short_url_id",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "location_not_found",
			expectedStatus: http.StatusBadRequest,
			shortURL:       "LFGwsFFf",
			expectedError:  "location not found",
			wantError:      true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(
			tt.name, func(t *testing.T) {
				urlList = sync.Map{}
				resp, respBody := testRequest(t, ts, "GET", "/"+tt.shortURL, "")
				defer resp.Body.Close()
				assert.Equal(t, tt.expectedStatus, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
				if tt.wantError {
					require.Contains(t, respBody, tt.expectedError, "Текст ошибки не совпадает с ожидаемым")
					require.NotEmpty(t, respBody, "Тело ответа пустое")
				} else {
					require.Empty(t, respBody, "Тело ответа не пустое")
				}
			},
		)
	}
}

func TestJSONHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		body           string
		baseURL        string
		wantError      bool
	}{
		{
			name:           "success_json_post_request",
			method:         http.MethodPost,
			expectedStatus: http.StatusCreated,
			body:           "{\n  \"url\": \"https://practicum.yandex.ru\"\n} ",
			baseURL:        "http://localhost:8080/",
		},
		{
			name:           "get_method",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
			wantError:      true,
		},
		{
			name:           "empty body",
			method:         http.MethodPost,
			expectedStatus: http.StatusBadRequest,
			wantError:      true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(
			tt.name, func(t *testing.T) {
				req := resty.New().R()
				req.Method = tt.method
				req.URL = ts.URL + "/api/shorten"
				req.SetHeader("Content-Type", "application/json")
				req.SetBody(tt.body)
				resp, err := req.Send()
				require.NoError(t, err)
				urlList = sync.Map{}
				t.Log(resp.StatusCode())
				assert.Equal(t, tt.expectedStatus, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")
				if !tt.wantError {
					require.NotEmpty(t, resp, "Тело ответа пустое")
					rb := resp.Body()
					require.True(t, strings.Contains(string(rb), tt.baseURL))
					require.True(t, len(resp.Body()) > len(tt.baseURL))
				}
			},
		)
	}
}
