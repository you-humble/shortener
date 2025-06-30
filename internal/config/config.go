package config

import (
	"flag"
	"os"
	"strings"
)

type Config struct {
	App App
}

type App struct {
	Host     string
	Port     string
	BaseAddr string
	LogLevel string
}

func (a App) Addr() string {
	return a.Host + ":" + a.Port
}

func MustLoad() *Config {
	const baseAddr string = "localhost:8080"

	var aAddr, bAddr, logLevel string
	flag.StringVar(&aAddr, "a", baseAddr, "HTTP server addres")
	flag.StringVar(&bAddr, "b", baseAddr, "base short URL address")
	flag.StringVar(&logLevel, "l", "info", "log level")

	flag.Parse()

	if envAAddr, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		aAddr = envAAddr
	}

	if envBAddr, ok := os.LookupEnv("BASE_URL"); ok {
		bAddr = envBAddr
	}

	if envLogLevel, ok := os.LookupEnv("LOG_LEVEL"); ok {
		logLevel = envLogLevel
	}

	hostPort := strings.Split(aAddr, ":")
	if len(hostPort) != 2 {
		panic("invalid app address: " + aAddr)
	}

	if !strings.Contains(bAddr, ":") {
		panic("invalid base address: " + aAddr)
	}

	cfg := new(Config)
	cfg.App.Host = hostPort[0]
	cfg.App.Port = hostPort[1]
	cfg.App.BaseAddr = bAddr
	cfg.App.LogLevel = logLevel

	return cfg
}
