package config

import (
	"os"
	"strconv"
	"time"
)

type Env string

const (
	Dev        Env = "dev"
	Prod       Env = "prod"
	defaultDSN     = "postgres://postgres:postgres@localhost:5432/phonebook?sslmode=disable"
)

type HTTP struct {
	Addr              string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int
}

type Postgres struct {
	URL      string // dsn -> postgres://user:pass@host:5432/db?sslmode=disable
	MaxConns int32
	MinConns int32
	Lifespan time.Duration
}

type Config struct {
	Env      Env
	HTTP     HTTP
	Postgres Postgres
}

func Load() Config {
	env := Env(getenv("APP_ENV", string(Dev)))
	http := HTTP{
		Addr:              getenv("HTTP_ADDR", ":8080"),
		ReadTimeout:       getdur("HTTP_READ_TIMEOUT", 5*time.Second),
		ReadHeaderTimeout: getdur("HTTP_READ_HEADER_TIMEOUT", 3*time.Second),
		WriteTimeout:      getdur("HTTP_WRITE_TIMEOUT", 10*time.Second),
		IdleTimeout:       getdur("HTTP_IDLE_TIMEOUT", 60*time.Second),
		MaxHeaderBytes:    getint("HTTP_MAX_HEADER_BYTES", 1<<20), // это равно 2^20 байтам, то есть 1MB
	}
	postgres := Postgres{
		URL:      getenv("PG_URL", defaultDSN),
		MaxConns: int32(getint("PG_MAX_CONNS", 50)),
		MinConns: int32(getint("PG_MIN_CONNS", 5)),
		Lifespan: getdur("PG_CONN_LIFESPAN", 30*time.Second),
	}
	return Config{
		Env:      env,
		HTTP:     http,
		Postgres: postgres,
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getdur(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return def
}

func getint(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		i, err := strconv.Atoi(v)
		if err == nil {
			return i
		}
	}
	return def
}
