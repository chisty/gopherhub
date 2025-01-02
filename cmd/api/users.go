package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/chisty/gopherhub/internal/store"
	"github.com/go-chi/chi/v5"
)

func (app *app) getUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("getUserHandler")

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user, err := app.store.Users.GetByID(r.Context(), id)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
	}
}
