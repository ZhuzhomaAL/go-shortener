package config

import (
	"flag"
	"os"
)

type AppConfig struct {
	FlagRunAddr   string
	FlagShortAddr string
	FlagLogLevel  string
	FlagStorage   string
	FlagDB        string
}

func ParseFlags() AppConfig {
	var appConfig AppConfig
	flag.StringVar(&appConfig.FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&appConfig.FlagShortAddr, "b", "http://localhost:8080", "address and port before short url")
	flag.StringVar(&appConfig.FlagLogLevel, "l", "info", "log level")
	flag.StringVar(&appConfig.FlagStorage, "f", "/tmp/short-url-db.json", "json file address")
	flag.StringVar(
		&appConfig.FlagDB, "d", "host=localhost user=postgres password=345973 dbname=postgres sslmode=disable", "database connection",
	)
	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		appConfig.FlagRunAddr = envRunAddr
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		appConfig.FlagShortAddr = envBaseURL
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		appConfig.FlagLogLevel = envLogLevel
	}

	if envStorage := os.Getenv("FILE_STORAGE_PATH"); envStorage != "" {
		appConfig.FlagStorage = envStorage
	}

	if envDB := os.Getenv("DATABASE_DSN"); envDB != "" {
		appConfig.FlagDB = envDB
	}

	return appConfig
}
