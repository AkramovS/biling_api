package main

import (
	"errors"
	"net/http"
	"time"

	"biling_api/internal/data"
	"biling_api/internal/validator"
)

// loginHandler handles user login and returns JWT token
func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	v.Check(input.Login != "", "login", "must be provided")
	v.Check(input.Password != "", "password", "must be provided")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.AuthUsers.Authenticate(input.Login, input.Password)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrInvalidCredentials):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Generate JWT token (24 hour expiry)
	token, err := app.models.Tokens.GenerateToken(user.ID, user.Login, 24*time.Hour)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{
		"token": token,
		"user":  user,
	}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
