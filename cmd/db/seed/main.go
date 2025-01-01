package main

import (
	"time"

	"github.com/chisty/gopherhub/internal/db"
	"github.com/chisty/gopherhub/internal/env"
	"github.com/chisty/gopherhub/internal/store"
)

func main() {
	addr := env.GetString("DB_ADDR", "postgres://postgres:postgres@localhost:5432/gopherhub?sslmode=disable")
	maxOpenConns := env.GetInt("DB_MAX_OPEN_CONNS", 20)
	maxIdleConns := env.GetInt("DB_MAX_IDLE_CONNS", 20)
	maxIdleTime := env.GetDuration("DB_MAX_IDLE_TIME", 10*time.Minute)

	conn, err := db.New(addr, maxOpenConns, maxIdleConns, maxIdleTime)
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	store := store.NewStorage(conn)

	db.Seed(store)
}
