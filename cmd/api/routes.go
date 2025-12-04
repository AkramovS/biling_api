package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// routes returns the router with all application routes
func (app *application) routes() http.Handler {
	router := httprouter.New()

	// Custom error handlers
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// Public routes
	router.HandlerFunc(http.MethodGet, "/v1/health", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/auth/register", app.registerHandler)
	router.HandlerFunc(http.MethodPost, "/v1/auth/login", app.loginHandler)

	// Protected routes - require authentication and permission
	router.HandlerFunc(http.MethodGet, "/v1/users/:id/accounts",
		app.requirePermission("accounts", "read", app.getUserAccountsHandler))

	return app.recoverPanic(app.enableCORS(router))
}
