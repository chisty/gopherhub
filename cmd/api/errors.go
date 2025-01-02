package main

import (
	"log"
	"net/http"
)

func (app *app) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("internal server error: %s path: %s error: %s\n ", r.Method, r.URL.Path, err.Error())
	writeJSONError(w, http.StatusInternalServerError, err.Error())
}

func (app *app) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("bad request error: %s path: %s error: %s\n ", r.Method, r.URL.Path, err.Error())
	writeJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *app) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("not found error: %s path: %s error: %s\n ", r.Method, r.URL.Path, err.Error())
	writeJSONError(w, http.StatusNotFound, err.Error())
}

func (app *app) conflictResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("conflict error: %s path: %s error: %s\n ", r.Method, r.URL.Path, err.Error())
	writeJSONError(w, http.StatusConflict, err.Error())
}
