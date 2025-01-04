package main

import (
	"log"
	"net/http"

	"github.com/chisty/gopherhub/internal/util"
)

type PaginatedFeedRequest struct {
	Limit  int    `json:"limit" validate:"gte=0,lte=100"`
	Offset int    `json:"offset" validate:"gte=0"`
	Sort   string `json:"sort" validate:"oneof=asc desc"`
}

func (app *app) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("getUserFeedHandler")

	fq, err := util.ParsePagination(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(fq); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	posts, err := app.store.Posts.GetUserFeed(r.Context(), int64(10), fq)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, posts); err != nil {
		app.internalServerError(w, r, err)
	}
}
