package app

import (
	"context"
	"github.com/ZhuzhomaAL/go-shortener/cmd/config"
	"github.com/ZhuzhomaAL/go-shortener/internal/logger"
	"github.com/ZhuzhomaAL/go-shortener/internal/store"
	"go.uber.org/zap"
	"time"
)

type app struct {
	appConfig config.AppConfig
	myLogger  logger.MyLogger
	reader    store.Reader
	writer    store.Writer
	storeChan chan store.URL
}

func NewApp(appConfig config.AppConfig, myLogger logger.MyLogger, reader store.Reader, writer store.Writer) *app {
	a := &app{appConfig: appConfig, myLogger: myLogger, reader: reader, writer: writer, storeChan: make(chan store.URL)}

	go a.deleteURLS()

	return a
}

func (a *app) deleteURLS() {
	ticker := time.NewTicker(10 * time.Second)

	var URLs []store.URL

	for {
		select {
		case URL := <-a.storeChan:
			URLs = append(URLs, URL)
		case <-ticker.C:
			if len(URLs) == 0 {
				continue
			}
			if writer, ok := a.writer.(store.WriterDeleter); ok {
				err := writer.DeleteURLs(context.Background(), URLs)
				if err != nil {
					a.myLogger.L.Error("failed to delete URLs", zap.Error(err))
					continue
				}
				a.myLogger.L.Info("successfully deleted URLs")
				URLs = nil
			}
		}
	}
}
