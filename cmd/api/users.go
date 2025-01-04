package main

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/chisty/gopherhub/internal/store"
	"github.com/chisty/gopherhub/internal/util"
	"github.com/go-chi/chi/v5"
)

type userKey string

const userKeyCtx userKey = "user"

type FollowUser struct {
	UserID int64 `json:"user_id" validate:"required"`
}

func (app *app) getUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("getUserHandler")

	user := getUserFromContext(r)

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *app) followUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("followUserHandler")

	// TODO: implement auth userID from context
	followerUser := getUserFromContext(r)
	var payload FollowUser
	if err := util.ReadJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := app.store.Followers.Follow(r.Context(), payload.UserID, followerUser.ID); err != nil {
		switch err {
		case store.ErrorConflictDuplicateKey:
			app.conflictResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *app) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("unfollowUserHandler")

	followerUser := getUserFromContext(r)

	// TODO: implement auth userID from context
	var payload FollowUser
	if err := util.ReadJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := app.store.Followers.UnFollow(r.Context(), payload.UserID, followerUser.ID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *app) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		ctx := context.WithValue(r.Context(), userKeyCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromContext(r *http.Request) *store.User {
	return r.Context().Value(userKeyCtx).(*store.User)
}
