package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/chisty/gopherhub/internal/store"
	"github.com/chisty/gopherhub/internal/util"
	"github.com/google/uuid"
)

type RegisterUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=4,max=40"`
}

// registerUserHandler godoc
//
//	@Summary		Register a user
//	@Description	Register a user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		RegisterUserRequest	true	"User credentials"
//	@Success		201		{object}	store.User			"User registered"
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

	if err := app.jsonResponse(w, http.StatusCreated, user); err != nil {
		app.internalServerError(w, r, err)
	}
}
