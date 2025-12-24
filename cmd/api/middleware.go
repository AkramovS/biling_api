package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"biling_api/internal/data"
)

// contextKey is a custom type for context keys
type contextKey string

const authUserContextKey = contextKey("authUser")

// authenticate validates JWT token and adds user to context
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			app.authenticationRequiredResponse(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]

		claims, err := app.models.Tokens.ValidateToken(token)
		if err != nil {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		user, err := app.models.AuthUsers.GetByLogin(claims.Login)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		ctx := context.WithValue(r.Context(), authUserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// requirePermission checks if authenticated user has required permission (fid)
func (app *application) requirePermission(fid int64, next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetAuthUser(r)
		if user == nil {
			app.authenticationRequiredResponse(w, r)
			return
		}

		hasPermission, err := app.models.Groups.HasPermission(user.ID, fid)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		if !hasPermission {
			app.notPermittedResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	}

	return app.authenticate(http.HandlerFunc(fn)).ServeHTTP
}

// contextGetAuthUser retrieves the AuthUser from the request context
func (app *application) contextGetAuthUser(r *http.Request) *data.AuthUser {
	user, ok := r.Context().Value(authUserContextKey).(*data.AuthUser)
	if !ok {
		return nil
	}

	return user
}

// enableCORS enables CORS for all requests
func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// recoverPanic recovers from panics and sends a 500 error
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
