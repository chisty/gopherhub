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

func (app *app) getPostHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("getPostHandler")

	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	post, err := app.store.Posts.GetByID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			writeJSONError(w, http.StatusNotFound, err.Error())
		default:
			writeJSONError(w, http.StatusInternalServerError, err.Error())
		}

		return
	}

	if err := writeJSON(w, http.StatusOK, post); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
	}
}
