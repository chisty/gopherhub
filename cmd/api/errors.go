package main

import (
	"net/http"

	"github.com/chisty/gopherhub/internal/util"
)

func (app *app) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("internal server error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
}

func (app *app) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("bad request error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	util.WriteJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *app) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("not found error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	util.WriteJSONError(w, http.StatusNotFound, err.Error())
}

func (app *app) conflictResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("conflict error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	util.WriteJSONError(w, http.StatusConflict, err.Error())
}

// func (app *app) unauthorizedErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
// 	app.logger.Errorw("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
// 	util.WriteJSONError(w, http.StatusUnauthorized, err.Error())
// }

func (app *app) unauthorizedBasicErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted", charset="UTF-8"`)

	util.WriteJSONError(w, http.StatusUnauthorized, err.Error())
}
