package main

import "log"

func main() {
	cfg := config{
		addr: ":8080",
	}

	app := app{
		config: cfg,
	}

	log.Fatal(app.run(app.mux()))

}
