package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/chisty/gopherhub/internal/mailer"
	"github.com/chisty/gopherhub/internal/store"
	"github.com/chisty/gopherhub/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RegisterUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=4,max=40"`
}

type UserWithToken struct {
	*store.User
	Token string `json:"token"`
}

// registerUserHandler godoc
//
//	@Summary		Register a user
//	@Description	Register a user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		RegisterUserRequest	true	"User credentials"
//	@Success		201		{object}	UserWithToken		"User registered"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/auth/user [post]
func (app *app) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Info("registerUserHandler")
	var registerUserPayload RegisterUserRequest
	if err := util.ReadJSON(w, r, &registerUserPayload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(registerUserPayload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &store.User{
		Username: registerUserPayload.Username,
		Email:    registerUserPayload.Email,
	}

	if err := user.Password.Set(registerUserPayload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// Generate a token for the user
	plainToken := uuid.New().String()
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	// Create and invite the user
	err := app.store.Users.CreateAndInvite(r.Context(), user, hashToken, app.config.mail.expiry)
	if err != nil {
		switch err {
		case store.ErrDuplicateEmail:
			app.badRequestResponse(w, r, err)
		case store.ErrDuplicateUsername:
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	userWithToken := &UserWithToken{
		User:  user,
		Token: plainToken,
	}

	activationURL := fmt.Sprintf("%x/confirm/%s", app.config.frontendURL, plainToken)
	isProdEnv := app.config.env == "production"

	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: activationURL,
	}

	err = app.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, !isProdEnv)
	if err != nil {
		app.logger.Errorw("error sending welcome email, rolling back user registration", "error", err)

		// rollback user creation if email fails (SAGA pattern)
		if err := app.store.Users.Delete(r.Context(), user.ID); err != nil {
			app.logger.Errorw("error rolling back user creation", "error", err)
		}

		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, userWithToken); err != nil {
		app.internalServerError(w, r, err)
	}
}

// activateUserHandler godoc
//
//	@Summary		Activate a user
//	@Description	Activate a user by invitation token
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			token	path		string	true	"Invitation token"
//	@Success		204		{string}	string	"User activated"
//	@Failure		400		{object}	error	"User payload missing"
//	@Failure		404		{object}	error	"User not found"
//	@Security		ApiKeyAuth
//	@Router			/users/activate/{token} [put]
func (app *app) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	err := app.store.Users.Activate(r.Context(), token)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}
