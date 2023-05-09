package config

import (
	"flag"
)

var (
	FlagRunAddr   string
	FlagShortAddr string
)

func ParseFlags() {
	flag.StringVar(&FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&FlagShortAddr, "b", "http://localhost:8080/", "address and port before short url")
	flag.Parse()
}
