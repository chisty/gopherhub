package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func (app *app) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("authorization header is required"))
			return
		}

		splits := strings.Split(authHeader, " ")
		if len(splits) != 2 || splits[0] != "Bearer" {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("invalid authorization header"))
			return
		}

		token := splits[1]
		jwtToken, err := app.authenticator.ValidateToken(token)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("invalid token"))
			return
		}

		claims, _ := jwtToken.Claims.(jwt.MapClaims)
		userID, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		user, err := app.store.Users.GetByID(r.Context(), userID)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		ctx := context.WithValue(r.Context(), userKeyCtx, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *app) AuthBasicMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
