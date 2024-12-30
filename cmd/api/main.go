package main

import (
	"log"
	"time"

	"github.com/chisty/gopherhub/internal/db"
	"github.com/chisty/gopherhub/internal/env"
	"github.com/chisty/gopherhub/internal/store"
)

func main() {
	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://postgres:postgres@localhost:5432/gopherhub?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 20),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 20),
			maxIdleTime:  env.GetDuration("DB_MAX_IDLE_TIME", 10*time.Minute),
		},
		env:     env.GetString("ENV", "development"),
		version: env.GetString("VERSION", "0.0.1"),
	}

	db, err := db.New(cfg.db.addr, cfg.db.maxOpenConns, cfg.db.maxIdleConns, cfg.db.maxIdleTime)
	if err != nil {
		log.Panic(err)
	}

	defer db.Close()
	log.Println("Database connection established")

	storage := store.NewStorage(db)

	app := app{
		config: cfg,
		store:  storage,
	}

	mux := app.mux()
	log.Fatal(app.run(mux))
}
