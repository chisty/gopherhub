package main

import (
	"log"
	"net/http"
)

func (app *app) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("getUserFeedHandler")

	posts, err := app.store.Posts.GetUserFeed(r.Context(), int64(10))
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, posts); err != nil {
		app.internalServerError(w, r, err)
	}
}
