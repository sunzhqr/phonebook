ifneq (,$(wildcard .env))
include .env
export $(shell sed -n 's/^\([A-Za-z_][A-Za-z0-9_]*\)=.*/\1/p' .env)
endif

.PHONY: run build migrate-up migrate-reset
run:
	go run ./cmd/server
build:
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o bin/server ./cmd/server
migrate-up:
	psql $$PG_URL -f db/migrations/0001_init.sql
migrate-reset:
	psql $$PG_URL -c "drop schema public cascade; create schema public;" && make migrate-up

test:
	 go test ./... -v

