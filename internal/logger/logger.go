package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type MyLogger struct {
	L *zap.Logger
}

func (l *MyLogger) Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	l.L = zl
	return nil
}

func (l *MyLogger) RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			responseData := &responseData{
				status: 0,
				size:   0,
			}
			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}
			h.ServeHTTP(&lw, r)
			l.L.Info(
				"got incoming HTTP request",
				zap.String("method", r.Method),
				zap.String("url", r.RequestURI),
				zap.Duration("latency", time.Since(start)),
				zap.Int("status", responseData.status),
				zap.Int("size", responseData.size),
			)
		},
	)
}

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
