package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/chisty/gopherhub/internal/store"
	"github.com/go-chi/chi/v5"
)

type CreatePostRequest struct {
	Title   string   `json:"title" validate:"required,max=200"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

func (app *app) createPostHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("createPostHandler")

	ctx := r.Context()

	var createPostPayload CreatePostRequest
	if err := readJSON(w, r, &createPostPayload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(createPostPayload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	post := &store.Post{
		Title:   createPostPayload.Title,
		Content: createPostPayload.Content,
		UserID:  1,
		Tags:    createPostPayload.Tags,
	}

	if err := app.store.Posts.Create(ctx, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := writeJSON(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *app) getPostHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("getPostHandler")

	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	post, err := app.store.Posts.GetByID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}

		return
	}

	comments, err := app.store.Comments.GetByPostID(r.Context(), post.ID)
	if err != nil {
		app.internalServerError(w, r, err)
	}

	post.Comments = comments

	if err := writeJSON(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
	}
}
