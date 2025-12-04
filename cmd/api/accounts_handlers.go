package main

import (
	"errors"
	"net/http"

	"biling_api/internal/data"
)

// getUserAccountsHandler returns all accounts for a user
func (app *application) getUserAccountsHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Verify user exists
	user, err := app.models.Users.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Get user's accounts
	accounts, err := app.models.Accounts.GetByUserID(id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{
		"user":     user,
		"accounts": accounts,
	}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
