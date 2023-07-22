package main

import (
	"fmt"
	"github.com/ZhuzhomaAL/go-shortener/cmd/config"
	"github.com/ZhuzhomaAL/go-shortener/internal/app"
	"github.com/ZhuzhomaAL/go-shortener/internal/file"
	"github.com/ZhuzhomaAL/go-shortener/internal/logger"
	"github.com/ZhuzhomaAL/go-shortener/internal/postgres"
	"github.com/ZhuzhomaAL/go-shortener/internal/store"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"sync"
)

var urlList sync.Map

func main() {
	appConfig := config.ParseFlags()
	var reader store.Reader
	var writer store.Writer

	switch {
	case appConfig.FlagDB != "":
		db := postgres.GetConnection(appConfig.FlagDB)
		defer db.Close()
		err := postgres.InitializeDB(db)
		if err != nil {
			log.Fatal(err)
		}
		reader = &store.DBReader{DB: db}
		writer = &store.DBWriter{DB: db}
	case appConfig.FlagStorage != "":
		urlList = sync.Map{}
		memoryReader := store.MemoryReader{
			UrlList: &urlList,
		}
		reader = &store.FileReader{MemoryReader: &memoryReader}
		memoryWriter := store.MemoryWriter{
			UrlList: &urlList,
		}
		fWriter, err := file.NewFileWriter(appConfig.FlagStorage)
		if err != nil {
			log.Fatal(err)
		}
		writer = &store.FileWriter{
			MemoryWriter: &memoryWriter, Writer: fWriter,
		}
		fReader, err := file.NewFileReader(appConfig.FlagStorage)
		if err != nil {
			log.Fatal(err)
		}
		for {
			url, err := fReader.ReadFile()
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Fatal(fmt.Errorf("failed to read the storage file: %w", err))
			}
			urlList.Store(url.ShortURL, url.OriginalURL)
		}
	default:
		urlList = sync.Map{}
		reader = &store.MemoryReader{
			UrlList: &urlList,
		}
		writer = &store.MemoryWriter{
			UrlList: &urlList,
		}
	}
	myLogger, err := logger.Initialize(appConfig.FlagLogLevel)
	if err != nil {
		log.Fatal(err)
	}
	a := app.NewApp(appConfig, myLogger, reader, writer)
	r, err := app.Router(a)
	if err != nil {
		log.Fatal(err)
	}
	myLogger.L.Info("Running server", zap.String("address", appConfig.FlagRunAddr))
	err = http.ListenAndServe(appConfig.FlagRunAddr, r)
	if err != nil {
		log.Fatal(err)
	}
}
