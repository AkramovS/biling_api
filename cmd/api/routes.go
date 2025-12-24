package main

import (
	"net/http"

	"biling_api/internal/data"

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
	router.HandlerFunc(http.MethodPost, "/v1/auth/login", app.loginHandler)

	// Protected routes
	router.HandlerFunc(http.MethodGet, "/v1/users/:id/accounts",
		app.requirePermission(data.FIDAccountsRead, app.getUserAccountsHandler))

	router.HandlerFunc(http.MethodGet, "/v1/account-tariffs/:id",
		app.requirePermission(data.FIDTariffsRead, app.getAccountTariffHandler))

	router.HandlerFunc(http.MethodPatch, "/v1/account-tariffs/:id",
		app.requirePermission(data.FIDTariffsUpdate, app.changeTariffLinkHandler))

	return app.recoverPanic(app.enableCORS(router))
}
