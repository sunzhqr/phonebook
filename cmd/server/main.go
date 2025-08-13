package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sunzhqr/phonebook/internal/config"
	"github.com/sunzhqr/phonebook/internal/httpserver"
	"github.com/sunzhqr/phonebook/internal/logger"
	"github.com/sunzhqr/phonebook/internal/repository"
	"github.com/sunzhqr/phonebook/internal/service"
)

func main() {
	cfg := config.Load()
	lg := logger.New(cfg.Env)
	defer lg.Sync()

	pool, err := repository.OpenDB(context.Background(), cfg.Postgres)
	if err != nil {
		lg.Fatal("db connect failed", logger.Err(err))
	}
	defer pool.Close()

	repos := repository.New(pool)
	svc := service.New(lg, repos.Contacts)
	httpSrv := httpserver.New(lg, cfg, svc)

	go func() {
		lg.Info("http listen", logger.KV("addr", cfg.HTTP.Addr))
		if err := httpSrv.Start(); err != nil {
			lg.Fatal("http server error", logger.Err(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpSrv.Stop(ctx)
	lg.Info("stopped")
}
