package config

import (
	"flag"
	"os"
)

var (
	FlagRunAddr   string
	FlagShortAddr string
)

func ParseFlags() {
	flag.StringVar(&FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&FlagShortAddr, "b", "http://localhost:8080", "address and port before short url")
	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		FlagRunAddr = envRunAddr
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		FlagShortAddr = envBaseURL
	}
}
