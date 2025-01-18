package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/chisty/gopherhub/internal/store"
	"github.com/chisty/gopherhub/internal/util"
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

// CreatePost godoc
//
//	@Summary		Creates a post
//	@Description	Creates a post
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreatePostRequest	true	"Post payload"
//	@Success		201		{object}	store.Post
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts [post]
func (app *app) createPostHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Info("createPostHandler")

	ctx := r.Context()

	var createPostPayload CreatePostRequest
	if err := util.ReadJSON(w, r, &createPostPayload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(createPostPayload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := getUserFromContext(r)

	post := &store.Post{
		Title:   createPostPayload.Title,
		Content: createPostPayload.Content,
		UserID:  user.ID,
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

// GetPost godoc
//
//	@Summary		Fetches a post
//	@Description	Fetches a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Post ID"
//	@Success		200	{object}	store.Post
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [get]
func (app *app) getPostHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Info("getPostHandler")

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
	app.logger.Info("deletePostHandler")

	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := app.store.Posts.Delete(r.Context(), id); err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdatePost godoc
//
//	@Summary		Updates a post
//	@Description	Updates a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Post ID"
//	@Param			payload	body		UpdatePostRequest	true	"Post payload"
//	@Success		200		{object}	store.Post
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [patch]
func (app *app) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromContext(r)

	var updatePostPayload UpdatePostRequest
	if err := util.ReadJSON(w, r, &updatePostPayload); err != nil {
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
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *app) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Info("createCommentHandler")

	post := getPostFromContext(r)

	var createCommentPayload CreateCommentRequest
	if err := util.ReadJSON(w, r, &createCommentPayload); err != nil {
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
			switch err {
			case store.ErrNotFound:
				app.notFoundResponse(w, r, err)
				return
			default:
				app.internalServerError(w, r, err)
				return
			}
		}

		ctx := context.WithValue(r.Context(), postKeyCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromContext(r *http.Request) *store.Post {
	return r.Context().Value(postKeyCtx).(*store.Post)
}
