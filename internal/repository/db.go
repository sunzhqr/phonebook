package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DBConfig - интерфейс для конфигурации для подключения
type DBConfig interface {
	GetURL() string
	GetMaxConns() int32
	GetMinConns() int32
}

// OpenDB открывает пул соединений pgxpool по URL из конфига.
func OpenDB(ctx context.Context, cfg DBConfig) (*pgxpool.Pool, error) {
	conf, err := pgxpool.ParseConfig(cfg.GetURL())
	if err != nil {
		return nil, err
	}
	conf.MaxConns = cfg.GetMaxConns()
	conf.MinConns = cfg.GetMinConns()
	return pgxpool.NewWithConfig(ctx, conf)
}
