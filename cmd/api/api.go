package main

import (
	"log"
	"net/http"
	"time"
)

type app struct {
	config config
}

type config struct {
	addr string
}

func (app *app) mux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/health", app.healthCheckHandler)
	return mux
}

func (app *app) run() error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      app.mux(),
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute * 1,
	}

	log.Println("Server is listening on", app.config.addr)

	return srv.ListenAndServe()
}
