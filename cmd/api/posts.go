package main

import (
	"log"
	"net/http"

	"github.com/chisty/gopherhub/internal/store"
)

type CreatePostRequest struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

func (app *app) createPostHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("createPostHandler")

	ctx := r.Context()

	var createPostPayload CreatePostRequest
	if err := readJSON(w, r, &createPostPayload); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	post := &store.Post{
		Title:   createPostPayload.Title,
		Content: createPostPayload.Content,
		UserID:  1,
		Tags:    createPostPayload.Tags,
	}

	if err := app.store.Posts.Create(ctx, post); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := writeJSON(w, http.StatusCreated, post); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
	}
}
