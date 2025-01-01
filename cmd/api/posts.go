package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/chisty/gopherhub/internal/store"
	"github.com/go-chi/chi/v5"
)

type postKey string

const postKeyCtx postKey = "post"

type CreatePostRequest struct {
	Title   string   `json:"title" validate:"required,max=200"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

type UpdatePostRequest struct {
	Title   *string `json:"title" validate:"omitempty,max=200"`
	Content *string `json:"content" validate:"omitempty,max=1000"`
}

type CreateCommentRequest struct {
	Content  string `json:"content" validate:"required,max=1000"`
	Username string `json:"username" validate:"required"`
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

	if err := app.jsonResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *app) getPostHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("getPostHandler")

	post := getPostFromContext(r)

	comments, err := app.store.Comments.GetByPostID(r.Context(), post.ID)
	if err != nil {
		app.internalServerError(w, r, err)
	}

	post.Comments = comments

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *app) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("deletePostHandler")

	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := app.store.Posts.Delete(r.Context(), id); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *app) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromContext(r)

	var updatePostPayload UpdatePostRequest
	if err := readJSON(w, r, &updatePostPayload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(updatePostPayload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if updatePostPayload.Title != nil {
		post.Title = *updatePostPayload.Title
	}

	if updatePostPayload.Content != nil {
		post.Content = *updatePostPayload.Content
	}

	if err := app.store.Posts.Update(r.Context(), post); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
	}

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *app) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("createCommentHandler")

	post := getPostFromContext(r)

	var createCommentPayload CreateCommentRequest
	if err := readJSON(w, r, &createCommentPayload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(createCommentPayload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user, err := app.store.Users.GetByUsername(r.Context(), createCommentPayload.Username)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	comment := &store.Comment{
		PostID:  post.ID,
		Content: createCommentPayload.Content,
		UserID:  user.ID,
	}

	if err := app.store.Comments.Create(r.Context(), comment); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, comment); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *app) postContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		ctx := context.WithValue(r.Context(), postKeyCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromContext(r *http.Request) *store.Post {
	return r.Context().Value(postKeyCtx).(*store.Post)
}
