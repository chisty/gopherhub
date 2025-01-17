package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

func (app *app) BasicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// read auth header
			// parse the header
			// decode
			// check if the user is valid

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.unauthorizedBasicErrorResponse(w, r, fmt.Errorf("authorization header is required"))
				return
			}

			splits := strings.Split(authHeader, " ")
			if len(splits) != 2 || splits[0] != "Basic" {
				app.unauthorizedBasicErrorResponse(w, r, fmt.Errorf("invalid authorization header"))
				return
			}

			decoded, err := base64.StdEncoding.DecodeString(splits[1])
			if err != nil {
				app.unauthorizedBasicErrorResponse(w, r, fmt.Errorf("failed to decode authorization header"))
				return
			}

			username := app.config.auth.basic.username
			password := app.config.auth.basic.password

			creds := strings.Split(string(decoded), ":")
			if len(creds) != 2 || creds[0] != username || creds[1] != password {
				app.unauthorizedBasicErrorResponse(w, r, fmt.Errorf("invalid credentials"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
