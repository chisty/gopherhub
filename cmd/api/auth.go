package main

import (
	"net/http"

	"github.com/chisty/gopherhub/internal/store"
	"github.com/chisty/gopherhub/internal/util"
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

	// hash the user password

	if err := user.Password.Set(registerUserPayload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, user); err != nil {
		app.internalServerError(w, r, err)
	}
}
