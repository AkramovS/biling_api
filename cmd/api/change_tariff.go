package main

import (
	"errors"
	"net/http"

	"biling_api/internal/data"
	"biling_api/internal/validator"
)

// getAccountTariffHandler returns current tariff assignment with metadata
// GET /v1/account-tariffs/:id
func (app *application) getAccountTariffHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	link, err := app.models.AccountTariffLinks.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Response includes version for optimistic locking
	err = app.writeJSON(w, http.StatusOK, envelope{
		"account_tariff": link,
	}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) changeTariffLinkHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Parse link ID from URL
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// 2. Parse request body
	var input struct {
		TariffID int64 `json:"tariff_id"`
		Version  int64 `json:"version"` // Expected version for optimistic locking
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// 3. Validate input
	v := validator.New()
	v.Check(input.TariffID > 0, "tariff_id", "must be a positive integer")
	v.Check(input.Version > 0, "version", "must be a positive integer")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// 4. Get current user from context (set by auth middleware)
	user := app.contextGetAuthUser(r)

	// 5. Prepare update with optimistic lock
	link := &data.AccountTariffLink{
		ID:        id,
		TariffID:  input.TariffID,
		Version:   input.Version,
		UpdatedBy: &user.ID,
	}

	// 6. Attempt update
	err = app.models.AccountTariffLinks.Update(link)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			// Version mismatch - get current server state
			app.editConflictResponse(w, r, id, input.TariffID, input.Version)
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// 7. Return updated record
	// Fetch full record with user info
	updatedLink, err := app.models.AccountTariffLinks.Get(id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{
		"account_tariff": updatedLink,
	}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// editConflictResponse returns 409 with both server and client data
// This allows UI to show what changed and preserve user input
func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request, id, clientTariffID, clientVersion int64) {
	// Get current server state
	serverLink, err := app.models.AccountTariffLinks.Get(id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	response := envelope{
		"error": map[string]interface{}{
			"code":    "version_conflict",
			"message": "Record was modified by another user. Please review changes and retry.",
			"details": map[string]interface{}{
				"entity": "account_tariff_link",
				"id":     id,
			},
		},
		// Current state on server - so UI can show what changed
		"server": map[string]interface{}{
			"data": map[string]interface{}{
				"id":         serverLink.ID,
				"account_id": serverLink.AccountID,
				"tariff_id":  serverLink.TariffID,
			},
			"meta": map[string]interface{}{
				"version":    serverLink.Version,
				"updated_at": serverLink.UpdatedAt,
				"updated_by": serverLink.UpdatedByUser,
			},
		},
		// What client tried to save - preserved for easy retry
		"client": map[string]interface{}{
			"data": map[string]interface{}{
				"tariff_id": clientTariffID,
			},
			"meta": map[string]interface{}{
				"expected_version": clientVersion,
			},
		},
	}

	err = app.writeJSON(w, http.StatusConflict, response, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
