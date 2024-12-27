package main

import (
	"log"

	"github.com/chisty/gopherhub/internal/env"
)

func main() {
	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
	}

	app := app{
		config: cfg,
	}

	log.Fatal(app.run(app.mux()))

}
