package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/chisty/gopherhub/internal/store"
	"github.com/go-chi/chi/v5"
)

type userKey string

const userKeyCtx userKey = "user"

type FollowUser struct {
	UserID int64 `json:"user_id" validate:"required"`
}

// GetUser          godoc
//
//	@Summary		etches a user profile
//	@Description	Fetches a user profile by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	store.User
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/{id} [get]
func (app *app) getUserHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Info("getUserHandler")

	userID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user, err := app.GetUser(r.Context(), userID)
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

// FollowUser          godoc
//
//	@Summary		Follow user
//	@Description	Follow a user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int		true	"User ID"
//	@Success		204	{string}	string	"User followed"
//	@Failure		400	{object}	error	"Bad Request"
//	@Failure		404	{object}	error	"User Not Found"
//	@Security		ApiKeyAuth
//	@Router			/users/{id}/follow [put]
func (app *app) followUserHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Info("followUserHandler")

	followerUser := getUserFromContext(r)
	followedID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := app.store.Followers.Follow(r.Context(), followerUser.ID, followedID); err != nil {
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

// UnfollowUser gdoc
//
//	@Summary		Unfollow a user
//	@Description	Unfollow a user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int		true	"User ID"
//	@Success		204		{string}	string	"User unfollowed"
//	@Failure		400		{object}	error	"User payload missing"
//	@Failure		404		{object}	error	"User not found"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/unfollow [put]
func (app *app) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Info("unfollowUserHandler")

	unfollowedUser := getUserFromContext(r)

	unfollowedID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := app.store.Followers.UnFollow(r.Context(), unfollowedUser.ID, unfollowedID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

//lint:ignore U1000 Will be used later
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

func (app *app) GetUser(ctx context.Context, userID int64) (*store.User, error) {
	if !app.config.redisCfg.enabled {
		return app.store.Users.GetByID(ctx, userID)
	}

	app.logger.Infow("fetching user from cache", "user_id", userID)

	user, err := app.cacheStore.Users.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		// fetch from db
		app.logger.Infow("cache miss! fetching user from db", "user_id", userID)

		user, err = app.store.Users.GetByID(ctx, userID)
		if err != nil {
			return nil, err
		}

		// set in cache
		if err := app.cacheStore.Users.Set(ctx, user); err != nil {
			return nil, err
		}
	}

	return user, nil
}
