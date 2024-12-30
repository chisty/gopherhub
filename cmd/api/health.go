package main

import (
	"log"
	"net/http"
)

func (app *app) healthCheckHandler(w http.ResponseWriter, r *http.Request) {

	data := map[string]string{
		"status":  "ok",
		"env":     app.config.env,
		"version": app.config.version,
	}

	if err := writeJSON(w, http.StatusOK, data); err != nil {
		log.Print("asdf")
		writeJSONError(w, http.StatusInternalServerError, err.Error())
	}
}
