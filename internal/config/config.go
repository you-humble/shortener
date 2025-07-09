package config

import (
	"flag"
	"os"
	"strings"
	"time"
)

type Config struct {
	App  App
	DB   Postgres
	Auth Auth
}

type App struct {
	Host     string
	Port     string
	BaseAddr string
	LogLevel string
}

type Postgres struct {
	DSN         string
	FileStorage string
}

type Auth struct {
	Secret      []byte
	TokenExpire time.Duration
}

func (a App) Addr() string {
	return a.Host + ":" + a.Port
}

func MustLoad() *Config {
	const (
		baseAddr      string = "localhost:8080"
		defaultFSPath string = "tmp/short-url-db.json"
	)

	var aAddr, bAddr, logLevel, fileStorage, dbDSN, secret string
	flag.StringVar(&aAddr, "a", baseAddr, "HTTP server addres")
	flag.StringVar(&bAddr, "b", baseAddr, "base short URL address")
	flag.StringVar(&logLevel, "l", "info", "log level")
	flag.StringVar(&fileStorage, "f", defaultFSPath, "file storage path")
	flag.StringVar(&dbDSN, "d", "", "database connection string")

	flag.Parse()

	if s, ok := os.LookupEnv("SECRET_KEY"); ok {
		secret = s
	}

	if envAAddr, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		aAddr = envAAddr
	}

	if envBAddr, ok := os.LookupEnv("BASE_URL"); ok {
		bAddr = envBAddr
	}

	if envLogLevel, ok := os.LookupEnv("LOG_LEVEL"); ok {
		logLevel = envLogLevel
	}

	if fs, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		fileStorage = fs
	}

	if db, ok := os.LookupEnv("DATABASE_DSN"); ok {
		// postgres://eblan:ne_eblan@localhost:5432/shortener?sslmode=disable
		dbDSN = db
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
	cfg.DB.FileStorage = fileStorage
	cfg.DB.DSN = dbDSN
	cfg.Auth.Secret = []byte(secret)
	cfg.Auth.TokenExpire = 24 * time.Hour

	return cfg
}
